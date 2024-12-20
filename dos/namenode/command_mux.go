package namenode

import (
	"log"

	"github.com/mrowaha/dos/api"
	"google.golang.org/grpc"
)

type DataNodeCommandMux struct {
	stream grpc.BidiStreamingServer[api.NodeHeartBeat, api.CommandNodeRes]
	logger *log.Logger
}

func NewDataNodeCommandMux(stream grpc.BidiStreamingServer[api.NodeHeartBeat, api.CommandNodeRes], logger *log.Logger) *DataNodeCommandMux {
	return &DataNodeCommandMux{
		stream: stream,
		logger: logger,
	}
}

func (mux *DataNodeCommandMux) Handle(cmd *CommandNode) {
	switch cmd.command {
	case api.CommandNodeRes_CREATE:
		mux.create(cmd.command, cmd.tag, &cmd.create)
	case api.CommandNodeRes_COMMIT:
		mux.commit(cmd.command, cmd.tag, &cmd.commit)
	case api.CommandNodeRes_DELETE:
		mux.delete(cmd.command, cmd.tag, &cmd.delete)
	case api.CommandNodeRes_UPDATE:
		mux.update(cmd.command, cmd.tag, &cmd.update)
	case api.CommandNodeRes_DISTRIBUTED_READ:
		mux.distributedRead(cmd.command, cmd.tag, &cmd.distributedRead)
	}
}

func (mux *DataNodeCommandMux) commit(cmd api.CommandNodeRes_Command, tag string, req *CommitCommand) {
	mux.logger.Printf("sending commit tagged %s @lamport%d\n", tag, req.Lamport)
	// time.Sleep(5 * time.Second)
	apicmd := &api.CommandNodeRes{
		Command:    cmd,
		MessageTag: tag,
		Commit: &api.CommitCommand{
			Lamport:    req.Lamport,
			ObjectName: req.Name,
		},
	}
	mux.stream.Send(apicmd)
}

func (mux *DataNodeCommandMux) create(cmd api.CommandNodeRes_Command, tag string, req *CreateCommand) {
	mux.logger.Printf("sending create tagged %s\n", tag)
	apicmd := &api.CommandNodeRes{
		Command:    cmd,
		MessageTag: tag,
		Create: &api.CreateCommand{
			ObjectName: req.Name,
			ObjectData: req.Data,
		},
	}
	mux.stream.Send(apicmd)
}

func (mux *DataNodeCommandMux) delete(cmd api.CommandNodeRes_Command, tag string, req *DeleteCommand) {
	mux.logger.Printf("sending delete tagged %s @lamport%d\n", tag, req.Lamport)
	apicmd := &api.CommandNodeRes{
		Command:    cmd,
		MessageTag: tag,
		Delete: &api.DeleteCommand{
			Lamport:    req.Lamport,
			ObjectName: req.Name,
		},
	}
	mux.stream.Send(apicmd)
}

func (mux *DataNodeCommandMux) update(cmd api.CommandNodeRes_Command, tag string, req *UpdateCommand) {
	mux.logger.Printf("sending update tagged %s @lamport%d\n", tag, req.Lamport)
	apicmd := &api.CommandNodeRes{
		Command:    cmd,
		MessageTag: tag,
		Update: &api.UpdateCommand{
			Lamport:    req.Lamport,
			ObjectName: req.Name,
			ObjectData: req.Data,
		},
	}
	mux.stream.Send(apicmd)
}

func (mux *DataNodeCommandMux) distributedRead(cmd api.CommandNodeRes_Command, tag string, req *DistributedReadCommand) {
	mux.logger.Printf("sending distributed read tagged %s @lamport%d\n", tag, req.Lamport)
	apicmd := &api.CommandNodeRes{
		Command:    cmd,
		MessageTag: tag,
		DistributedRead: &api.DistributedReadCommand{
			Objects: req.Objects,
			Lamport: req.Lamport,
		},
	}
	mux.stream.Send(apicmd)
}
