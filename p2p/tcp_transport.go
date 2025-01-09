package p2p

import (
	"fmt"
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
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.listnerAdder)
	if err != nil {
		return err
	}
	fmt.Printf("Listening on %v\n", t.listner.Addr())
	go t.StartAcceptingLoop()
	return nil
}

func (t *TCPTransport) StartAcceptingLoop() error {
	for {
		conn, err := t.listner.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			return err
		}
		go t.HandleConnection(conn)

	}
}
func (t *TCPTransport) HandleConnection(conn net.Conn) {
	fmt.Printf("New incoming connection from %s\n", conn.RemoteAddr())
}
