package memfs

import (
	"bytes"
	"io"
	"libnfs-go/fs"
	"sync"
)

type dataNode struct {
	id   uint64
	data []byte
}

func (n *dataNode) Id() uint64 {
	return n.id
}

func (n *dataNode) Reader() io.Reader {
	return bytes.NewReader(n.data)
}

func (n *dataNode) Size() int {
	return len(n.data)
}

type Storage struct {
	nodes  []*dataNode
	nodeId uint64
	lck    *sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		nodes: []*dataNode{},
		lck:   &sync.RWMutex{},
	}
}

func (s *Storage) nextId() uint64 {
	s.lck.Lock()
	defer s.lck.Unlock()
	s.nodeId++
	return s.nodeId
}

func (s *Storage) Create(src io.Reader) (uint64, error) {
	id := s.nextId()

	s.lck.Lock()
	defer s.lck.Unlock()

	data, err := io.ReadAll(src)
	if err != nil {
		if err != io.EOF {
			return id, err
		}
	}

	node := &dataNode{id: id, data: data}
	s.nodes = append(s.nodes, node)

	return id, nil
}

func (s *Storage) Update(id uint64, src io.Reader) (bool, error) {
	s.lck.Lock()
	defer s.lck.Unlock()

	for _, n := range s.nodes {
		if n.id == id {
			data, err := io.ReadAll(src)
			if err != nil {
				if err != io.EOF {
					return false, err
				}
			}
			n.data = data
			return true, nil
		}
	}

	return false, nil
}

func (s *Storage) Delete(id uint64) bool {
	s.lck.Lock()
	defer s.lck.Unlock()

	found := false
	for _, n := range s.nodes {
		if n.id == id {
			found = true
			break
		}
	}

	if found {
		nList := []*dataNode{}
		for _, n := range s.nodes {
			if n.id != id {
				nList = append(nList, n)
			}
		}
		s.nodes = nList
	}

	return found
}

func (s *Storage) Size(id uint64) int {
	n := s.Get(id)
	if n != nil {
		return n.Size()
	}
	return 0
}

func (s *Storage) Get(id uint64) fs.StorageNode {
	s.lck.RLock()
	defer s.lck.RUnlock()

	for _, n := range s.nodes {
		if n.id == id {
			return n
		}
	}

	return nil
}
