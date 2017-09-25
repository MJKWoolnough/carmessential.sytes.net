package main

import (
	"os"
	"sync"
	"time"

	"github.com/MJKWoolnough/httpdir"
)

type node struct {
	mu sync.RWMutex
	httpdir.Node
}

func (n *node) Size() int64 {
	n.mu.RLock()
	l := n.Node.Size()
	n.mu.RUnlock()
	return l
}

func (n *node) Mode() os.FileMode {
	return httpdir.ModeFile
}

func (n *node) ModTime() time.Time {
	n.mu.RLock()
	t := n.Node.Mode()
	n.mu.RUnlock()
	return t
}

func (n *node) Open() (httpdir.File, error) {
	n.mu.RLock()
	f, err := n.Node.Open()
	n.mu.RUnlock()
	return f, err
}

func (n *node) Update(data []byte, t time.Time) {
	n.mu.Lock()
	n.Node = httpdir.FileBytes(data, t)
	n.mu.Unlock()
}
