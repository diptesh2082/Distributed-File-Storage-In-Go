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
//
//	func (p *TCPPeer) String() string {
//		return p.conn.RemoteAddr().String() // This will print the remote address of the connection
//	}
type TCPTransportOpts struct {
	ListnerAdder  string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
}
type TCPTransport struct {
	TCPTransportOpts
	listner net.Listener
	mu      sync.RWMutex
	peers   map[net.Addr]Peer
}

func NewTCPPeer(conn net.Conn, outbount bool) *TCPPeer {

	return &TCPPeer{
		conn:     conn,
		outbount: outbount,
	}
}

// type Temp struct{}

// NewTCPTransport creates a new TCPTransport instance.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {

	return &TCPTransport{
		TCPTransportOpts: opts,
		peers:            make(map[net.Addr]Peer),
		mu:               sync.RWMutex{},
	}
}
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.ListnerAdder)
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
		// if err = t.HandShakeFunc(conn) ; err != nil{

		// }
		// fmt.Println("Accepted a new connection")
		go t.HandleConnection(conn)

	}
}
func (t *TCPTransport) HandleConnection(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("New incoming connection from %v\n", peer)

	if err := t.HandShakeFunc(conn); err != nil {
		fmt.Printf("Handshake error for connection from %v: %s\n", peer.conn.LocalAddr(), err)
		conn.Close()
		return
	}
	rpc := &RPC{}
	for {
		if err := t.Decoder.Decode(conn, rpc); err != nil {
			fmt.Printf("Error decoding message from %v: %s\n", peer.conn.LocalAddr(), err)
			// conn.Close()
			continue
		}
		rpc.From = conn.RemoteAddr()
		fmt.Printf("message :: %v\n", rpc)
	}
}
