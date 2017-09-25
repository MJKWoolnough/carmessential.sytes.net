package main

import (
	"compress/gzip"
	"os"
	"sync"
	"time"

	"github.com/MJKWoolnough/httpdir"
	"github.com/MJKWoolnough/memio"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriterLevel(nil, gzip.BestCompression)
	},
}

type nodes struct {
	httpdir.Dir

	mu    sync.Mutex
	nodes map[string]*node
}

func (n *nodes) Set(filename string, data []byte) {
	var gzipped memio.Buffer
	gw := gzipPool.Get().(*gzip.Writer)
	gw.Reset(&gzipped)
	gw.Write(data)
	gw.Close()
	gzipPool.Put(gw)
	t := time.Now()
	n.mu.Lock()
	nd, ok := n.nodes[filename]
	if !ok {
		nd = new(node)
		n.Create(filename, nd)
		n.nodes[filename] = nd
	}
	node.Update(data, t)
	if node, ok := n.nodes[filename+".gz"]; !ok {

	} else {
		node.Update(data, time.Now())
	}
	n.mu.Unlock()
}

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
