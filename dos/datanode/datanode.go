package datanode

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mrowaha/dos/api"
	"github.com/mrowaha/dos/namenode"
	"google.golang.org/grpc"
)

type DosDataNode struct {
	me     string
	client api.DataServiceClient
	logger *log.Logger
	store  *DataNodeSqlStore
	config *DataNodeConfig
	queue  DataNodeQueue
	leaser *DataNodeLeaseService
}

func NewDosDataNode(conn *grpc.ClientConn, queue DataNodeQueue, opts ...DNodeConfigFunc) *DosDataNode {
	c := api.NewDataServiceClient(conn)
	logger := log.New(os.Stdout, "[datanode]", log.Ltime)

	cfg := defaultDataNodeConfig()
	for _, fn := range opts {
		fn(cfg)
	}

	if len(cfg.leaserAddr) == 0 {
		log.Fatalln("node config error: leaserAddr cannot be empty")
	}

	if len(cfg.name) == 0 {
		log.Fatalln("node config error: name cannot be empty")
	}

	store := NewDataNodeSqlStore(cfg.dbFile)
	store.BootStrap()
	leaser := NewDataNodeLeaseService(cfg.leaserAddr)

	return &DosDataNode{
		client: c,
		logger: logger,
		config: cfg,
		store:  store,
		me:     cfg.name,
		queue:  queue,
		leaser: leaser,
	}
}

func (d *DosDataNode) HandleCommit(cmd *api.CommitCommand) {
	commit := false
	for create, _ := d.queue.PullCreateCmd(cmd.ObjectName); create != nil; create, _ = d.queue.PullCreateCmd(cmd.ObjectName) {
		commit = true
		sequence, err := d.store.Write(create.Name, create.Data)
		if err != nil {
			log.Fatalf("failed to flush create requests %v", err)
		}
		log.Printf("flushed create requests")
		d.leaser.PushObjectCreate(create.Name, create.Data, sequence)
	}
	if !commit {
		log.Printf("nothing to commit")
	}
}

func (d *DosDataNode) HandleDelete(cmd *api.DeleteCommand) {
	log.Printf("executing delete command for %s\n", cmd.ObjectName)
	sequence, err := d.store.Delete(cmd.ObjectName)
	if err != nil {
		if err == ErrObjectNotInStore {
			return
		}
		log.Fatalf("failed to handle delete object...\n%s\n", err.Error())
	}
	d.leaser.PushObjectDelete(cmd.ObjectName, sequence+1)
}

func (d *DosDataNode) HandleUpdate(cmd *api.UpdateCommand) {
	log.Printf("update object request %s\n", cmd.ObjectName)
	sequence, err := d.store.Update(cmd.ObjectName, cmd.ObjectData)
	if err != nil {
		if err == ErrObjectNotInStore {
			return
		}
		log.Fatalf("failed to handle update object...\n%s\n", err.Error())
	}
	d.leaser.PushObjectUpdate(cmd.ObjectName, cmd.ObjectData, sequence)
}

func (d *DosDataNode) Register() {
	bistream, err := d.client.RegisterNode(context.Background())
	if err != nil {
		log.Fatalf("failed to start register node\n%s", err.Error())
	}

	messageChan := make(chan *api.NodeHeartBeat, 5)

	go func() {
		for {
			// heartbeat
			size, _ := d.store.Size()
			objects, _ := d.store.Objects()
			messageChan <- &api.NodeHeartBeat{
				Id:            d.me,
				Size:          size,
				Objects:       objects,
				LeaserService: d.config.leaserAddr,
				Type:          api.NodeHeartBeat_BEAT,
			}
			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		for message := range messageChan {
			bistream.Send(message)
		}
	}()

	var lastLamport int32
	var blocked bool
	var skip bool
	lastLamport = 0 // have not yet received any lamport
	for {
		log.Printf("awaiting")
		blocked = false
		skip = false
		resp, err := bistream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				log.Fatalf("stream closed by server")
			}
			log.Fatalf("error receiving from stream: %v", err)
		}

		switch resp.Command {
		case api.CommandNodeRes_CREATE:
			log.Printf("received create command %s\n", resp.Create.ObjectName)
			// we are going to wait for commit
			_ = d.queue.PushCreateCmd(namenode.CreateCommand{
				Name: resp.Create.ObjectName,
				Data: resp.Create.ObjectData,
			})
			skip = true
			messageChan <- &api.NodeHeartBeat{
				Type:       api.NodeHeartBeat_ACK,
				MessageTag: resp.MessageTag,
			}
		case api.CommandNodeRes_COMMIT:
			// flush the create queue if commit contains this datanode
			log.Printf("received commit command, lamport %d\n", resp.Commit.Lamport)
			if resp.Commit.Lamport-lastLamport == 1 {
				lastLamport = resp.Commit.Lamport
				d.HandleCommit(resp.Commit)
			} else {
				log.Printf("commit command was blocked")
				blocked = true
				err := d.queue.BlockCommand(
					float64(resp.Commit.Lamport),
					namenode.CommitCommand{
						Type: namenode.COMMIT,
					},
				)
				if err != nil {
					log.Fatalf("failed to block command... %v\n", err.Error())
				}
			}
			messageChan <- &api.NodeHeartBeat{
				Type:       api.NodeHeartBeat_ACK,
				MessageTag: resp.MessageTag,
			}
		case api.CommandNodeRes_DELETE:
			log.Printf("delete object request, lamport %d\n", resp.Delete.Lamport)
			if resp.Delete.Lamport-lastLamport == 1 {
				lastLamport = resp.Delete.Lamport
				d.HandleDelete(resp.Delete)
			} else {
				log.Printf("delete command was blocked")
				blocked = true
				err := d.queue.BlockCommand(
					float64(resp.Delete.Lamport),
					namenode.DeleteCommand{
						Name: resp.Delete.ObjectName,
						Type: namenode.DELETE,
					},
				)
				if err != nil {
					log.Fatalf("failed to block delete command... %v\n", err.Error())
				}
			}
			messageChan <- &api.NodeHeartBeat{
				Type:       api.NodeHeartBeat_ACK,
				MessageTag: resp.MessageTag,
			}
		case api.CommandNodeRes_UPDATE:
			log.Printf("update object request %s, @lamport%d\n", resp.Update.ObjectName, resp.Update.Lamport)
			if resp.Update.Lamport-lastLamport == 1 {
				lastLamport = resp.Update.Lamport
				d.HandleUpdate(resp.Update)
			} else {
				log.Printf("update command was blocked")
				blocked = true
				err := d.queue.BlockCommand(
					float64(resp.Update.Lamport),
					namenode.UpdateCommand{
						Name: resp.Update.ObjectName,
						Data: resp.Update.ObjectData,
						Type: namenode.UPDATE,
					},
				)
				if err != nil {
					log.Fatalf("failed to block update command... %v\n", err.Error())
				}
			}
			messageChan <- &api.NodeHeartBeat{
				Type:       api.NodeHeartBeat_ACK,
				MessageTag: resp.MessageTag,
			}
		}

		// now we begin delivering all the messages and update the lastlamport accordingly
		// first we retrieve the command that had the smallest lamport value in the queue
		// if the diff between that lamport and current lamport is 1, deliver and update current
		// lamport
		if !skip && !blocked {

			for {
				next, lamport, err := d.queue.RetrieveCommand()
				if err == ErrNoCommandToRetrieve {
					break
				}
				if err != nil {
					log.Fatalf("%v\n", err)
				}
				if lamport-float64(lastLamport) != 1 {
					break
				}

				// now we know that this event is a valid blocked next
				lastLamport = int32(lamport)
				log.Printf("going to deliver command %s @lamport%.0f\n", next, lamport)
				next, _, err = d.queue.DeliverCommand()
				if err != nil {
					log.Fatalf("%s\n", err.Error())
				}
				switch v := next.(type) {
				case namenode.CommitCommand:
					d.HandleCommit(&api.CommitCommand{})
					log.Printf("delievered")
				case namenode.DeleteCommand:
					d.HandleDelete(&api.DeleteCommand{
						ObjectName: v.Name,
					})
					log.Printf("delievered")
				case namenode.UpdateCommand:
					d.HandleUpdate(&api.UpdateCommand{
						ObjectName: v.Name,
						ObjectData: v.Data,
					})
					log.Printf("delievered")
				}
			}

		}
	}
}
