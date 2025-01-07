package p2p

import (
	"net"
	"sync"
)

type TCPTransport struct {
	listnerAdder string
	listner      net.Listener
	mu           sync.RWMutex
	peers        map[net.Addr]Peer
}

// NewTCPTransport creates a new TCPTransport instance.
func NewTCPTransport(listnerAddr string) *TCPTransport {

	return &TCPTransport{
		listnerAdder: listnerAddr,
		peers:        make(map[net.Addr]Peer),
	}
}
