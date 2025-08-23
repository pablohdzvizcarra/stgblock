package client

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	ID          string
	Version     byte
	Addr        string
	Conn        net.Conn
	ConnectedAt time.Time
	Metadata    map[string]string
}

type ClientRegistry struct {
	mu    sync.RWMutex
	peers map[string]*Client
}

func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{peers: make(map[string]*Client)}
}

func (r *ClientRegistry) Add(p *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers[p.ID] = p
}

func (r *ClientRegistry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, id)
}

func (r *ClientRegistry) Get(id string) (*Client, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.peers[id]
	return p, ok
}
