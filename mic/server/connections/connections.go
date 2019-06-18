package connections

import (
	"net"
	"sync"
)

type Connections struct {
	connections map[uint64]net.Conn
	mu          *sync.RWMutex
	nextId      uint64
}

func (this *Connections) Add(conn net.Conn) uint64 {
	this.mu.Lock()
	defer this.mu.Unlock()
	id := this.nextId
	this.nextId++
	this.connections[id] = conn
	return id
}

func (this *Connections) ForEach(fn func(net.Conn, uint64)) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for id, conn := range this.connections {
		fn(conn, id)
	}
}

func (this *Connections) Remove(ids ...uint64) {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, id := range ids {
		delete(this.connections, id)
	}
}

func New() *Connections {
	return &Connections{
		mu:          new(sync.RWMutex),
		connections: make(map[uint64]net.Conn),
	}
}
