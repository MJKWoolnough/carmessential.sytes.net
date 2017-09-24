package main

import (
	"bytes"
	"os"
	"sync"
	"time"

	"github.com/MJKWoolnough/httpdir"
)

type Node struct {
	mu      sync.RWMutex
	Data    []byte
	Updated time.Time
}

func (n *NodeNode) Size() int64 {
	n.mu.RLock()
	l := int64(len(n.Data))
	n.mu.RUnlock()
	return l
}

func (n *Node) Mode() os.FileMode {
	return os.FileMode
}

func (n *Node) ModTime() time.Time {
	n.mu.RLock()
	t := n.Updated
	n.mu.RUnlock()
	return t
}

func (n *Node) Open() (httpdir.File, error) {
	n.mu.RLock()
	f := bytes.NewBuffer(n.Data)
	n.mu.RUnlock()
	return f, nil
}

func (n *Node) Update(data []byte, t time.Time) {
	n.mu.Lock()
	n.Data = data
	n.Updated = t
	n.mu.Unlock()
}
