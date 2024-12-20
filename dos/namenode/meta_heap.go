package namenode

import (
	"container/heap"
	"slices"
)

type MetaHeapEntry struct {
	Id        string
	Lease     string
	CommandCh chan<- CommandNode
	ResChs    map[string]chan interface{}
	ClosedCh  <-chan struct{}
	Size      float32
}

type MetaHeap []*MetaHeapEntry

func (h *MetaHeap) Len() int {
	return len(*h)
}

func (h *MetaHeap) Less(i, j int) bool {
	a, b := (*h)[i], (*h)[j]
	return a.Size > b.Size
}

func (h *MetaHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *MetaHeap) Push(x interface{}) {
	item := x.(*MetaHeapEntry)
	*h = append(*h, item)
}

func (h *MetaHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[0 : n-1]
	return item
}

// concurrent meta heap is a wrapper around meta heap providing synced access
type DataNodeMeta struct {
	heap *MetaHeap
}

func NewDataNodeMeta() *DataNodeMeta {
	metaHeap := &MetaHeap{}
	heap.Init(metaHeap)
	return &DataNodeMeta{
		heap: metaHeap,
	}
}

func (d *DataNodeMeta) RegisterNode(entry *MetaHeapEntry) {
	heap.Push(d.heap, entry)
}

func (d *DataNodeMeta) Count() int {
	return d.heap.Len()
}

func (d *DataNodeMeta) Next() *MetaHeapEntry {
	if d.heap.Len() == 0 {
		return nil
	}
	return heap.Pop(d.heap).(*MetaHeapEntry)
}

func (d *DataNodeMeta) Exists(id string) bool {
	for _, entry := range *d.heap {
		if entry.Id == id {
			return true
		}
	}
	return false
}

func (d *DataNodeMeta) ForEach(cb func(entry *MetaHeapEntry)) {
	for _, entry := range *d.heap {
		cb(entry)
	}
}

func (d *DataNodeMeta) UpdateSize(id string, size float32) bool {
	idx := slices.IndexFunc(*d.heap, func(c *MetaHeapEntry) bool {
		return c.Id == id
	})

	if (*d.heap)[idx].Size == size {
		return false
	}
	(*d.heap)[idx].Size = size
	heap.Fix(d.heap, idx)
	return true
}

// this function will return the lease services for the corresponding nodes
func (d *DataNodeMeta) LeaseServices(nodes []string) ([]string, error) {
	services := make([]string, 0)
	for _, el := range *d.heap {
		i := slices.IndexFunc(nodes, func(node string) bool {
			return el.Id == node
		})
		if i != -1 {
			//this el is in the nodes slice
			services = append(services, el.Lease)
		}
	}
	return services, nil
}

func (d *DataNodeMeta) DeleteNode(nodeId string) {
	idx := slices.IndexFunc(*d.heap, func(entry *MetaHeapEntry) bool {
		return entry.Id == nodeId
	})

	if idx == -1 {
		return
	}

	lastIdx := d.heap.Len() - 1
	(*d.heap)[idx] = (*d.heap)[lastIdx]
	*d.heap = (*d.heap)[:lastIdx]
	if idx < d.heap.Len() {
		heap.Fix(d.heap, idx)
	}
}
