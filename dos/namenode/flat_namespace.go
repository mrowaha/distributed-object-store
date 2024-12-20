package namenode

import (
	"bufio"
	"errors"
	"os"
	"slices"
)

/**
	this file creates the implementation for the flat namespace
	accesses to the flat namespace are single threaded and are synced by a global lock
**/

var (
	ErrObjectAlreadyExists = errors.New("object already exists")
	ErrLoadFlatNamespace   = errors.New("failed to load flat ns")
	ErrObjectDoestNotExist = errors.New("object does not exist")
)

type FlatNamespaceEntry struct {
	name  string
	nodes map[string]bool
}

type FlatNamespace struct {
	ns []FlatNamespaceEntry
}

func NewFlatNamespace() *FlatNamespace {
	return &FlatNamespace{
		ns: make([]FlatNamespaceEntry, 0),
	}
}

func (fn *FlatNamespace) Exists(name string) bool {
	for _, object := range fn.ns {
		if object.name == name {
			return true
		}
	}
	return false
}

func (fn *FlatNamespace) AddObject(name string) error {
	fn.ns = append(fn.ns, FlatNamespaceEntry{
		name:  name,
		nodes: make(map[string]bool),
	})

	return nil
}

func (fn *FlatNamespace) DeleteObject(name string) error {
	newNs := make([]FlatNamespaceEntry, 0)
	found := false
	for _, object := range fn.ns {
		if object.name == name {
			found = true
		} else {
			// if it is the not the object then append to new ns
			newNs = append(newNs, object)
		}
	}
	if found {
		fn.ns = newNs
		return nil
	}
	return ErrObjectDoestNotExist
}

func (fn *FlatNamespace) AddNode(forObject string, nodeId string) {
	for _, object := range fn.ns {
		if object.name == forObject {
			object.nodes[nodeId] = true
			return
		}
	}
}

func (fn *FlatNamespace) RemoveNode(nodeId string) {
	for _, object := range fn.ns {
		delete(object.nodes, nodeId)
	}
}

func (fn *FlatNamespace) Nodes(forObject string) ([]string, error) {
	i := slices.IndexFunc(fn.ns, func(e FlatNamespaceEntry) bool {
		return e.name == forObject
	})
	if i == -1 {
		return nil, ErrObjectDoestNotExist
	}

	nodes := make([]string, len(fn.ns[i].nodes))
	for k := range fn.ns[i].nodes {
		nodes = append(nodes, k)
	}
	return nodes, nil
}

func (fn *FlatNamespace) Load(fnFile string) error {
	f, err := os.OpenFile(fnFile, os.O_RDONLY, 0666)
	if err != nil {
		return ErrLoadFlatNamespace
	}
	defer f.Close()
	fn.ns = make([]FlatNamespaceEntry, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fn.ns = append(fn.ns, FlatNamespaceEntry{
			name: scanner.Text(),
		})
	}
	return nil
}
