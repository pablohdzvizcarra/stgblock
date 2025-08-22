package server

import (
	"net"
	"sync"
	"time"
)

type Peer struct {
	ID          string
	Version     byte
	Addr        string
	Conn        net.Conn
	ConnectedAt time.Time
	Metadata    map[string]string
}

type Registry struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

func NewRegistry() *Registry {
	return &Registry{peers: make(map[string]*Peer)}
}

func (r *Registry) Add(p *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers[p.ID] = p
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, id)
}

func (r *Registry) Get(id string) (*Peer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.peers[id]
	return p, ok
}
