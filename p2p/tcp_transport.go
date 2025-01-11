package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPPeer struct {
	conn     net.Conn
	outbount bool
}

// // Add a String method to TCPPeer
// func (p *TCPPeer) String() string {
// 	return p.conn.RemoteAddr().String() // This will print the remote address of the connection
// }

type TCPTransport struct {
	listnerAdder string
	listner      net.Listener
	mu           sync.RWMutex
	peers        map[net.Addr]Peer
}

func NewTCPPeer(conn net.Conn, outbount bool) *TCPPeer {

	return &TCPPeer{
		conn:     conn,
		outbount: outbount,
	}
}

// NewTCPTransport creates a new TCPTransport instance.
func NewTCPTransport(listnerAddr string) *TCPTransport {

	return &TCPTransport{
		listnerAdder: listnerAddr,
		peers:        make(map[net.Addr]Peer),
		mu:           sync.RWMutex{},
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
		// fmt.Println("Accepted a new connection")
		conn, err := t.listner.Accept()
		// fmt.Println("Accepted a new connection")
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			// return err
		}
		// fmt.Println("Accepted a new connection")
		go t.HandleConnection(conn)

	}
}
func (t *TCPTransport) HandleConnection(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("New incoming connection from %v\n", peer)
}
