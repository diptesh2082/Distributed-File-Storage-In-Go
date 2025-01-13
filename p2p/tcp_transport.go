package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn
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
	OnPeer        func(Peer) error
}
type TCPTransport struct {
	TCPTransportOpts
	listner net.Listener
	mu      sync.RWMutex
	rpcch   chan RPC
	// peers   map[net.Addr]Peer
}

func NewTCPPeer(conn net.Conn, outbount bool) *TCPPeer {

	return &TCPPeer{
		Conn:     conn,
		outbount: outbount,
	}
}

// type Temp struct{}

// NewTCPTransport creates a new TCPTransport instance.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {

	return &TCPTransport{
		TCPTransportOpts: opts,
		// peers:            make(map[net.Addr]Peer),
		mu:    sync.RWMutex{},
		rpcch: make(chan RPC, 1018),
	}
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (p *TCPTransport) Close() error {
	return p.listner.Close()
}

// func (p *TCPPeer) Close() error {
// 	return p.conn.Close()
// }

// func (p *TCPPeer) RemoteAdder() net.Addr {
// 	return p.conn.RemoteAddr()
// }

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.ListnerAdder)
	if err != nil {
		return err
	}

	fmt.Printf("Listening on %v\n", t.listner.Addr().String())
	go t.StartAcceptingLoop()
	return nil
}

func (t *TCPTransport) StartAcceptingLoop() error {
	for {
		// fmt.Println("Accepted a new connection")
		conn, err := t.listner.Accept()
		// fmt.Println("Accepted a new connection")
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
			// return err
		}
		// if err = t.HandShakeFunc(conn) ; err != nil{

		// }
		// fmt.Println("Accepted a new connection")
		go t.HandleConnection(conn, false)

	}
}
func (t *TCPTransport) HandleConnection(conn net.Conn, outbound bool) {
	var err error
	// fmt.Printf("New incoming connection from %v to %v\n", conn.RemoteAddr(), conn.LocalAddr())
	defer func() {
		fmt.Printf("dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err := t.HandShakeFunc(conn); err != nil {
		fmt.Printf("Handshake error for connection from %v: %s\n", peer.Conn.LocalAddr(), err)
		conn.Close()
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return
		}
	}

	for {
		rpc := &RPC{}
		// fmt.Printf("Raw payload bytes: %v\n", rpc.Payload)

		if err := t.Decoder.Decode(conn, rpc); err != nil {
			fmt.Printf("Error decoding message from %v: %s\n", peer.Conn.LocalAddr(), err)
			conn.Close()
			continue
		}
		rpc.From = conn.RemoteAddr()
		fmt.Printf("Raw payload bytes: %v\n", rpc.Payload)
		fmt.Printf("message :: %v\n %s", rpc, conn.RemoteAddr())
		t.rpcch <- *rpc
	}
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.HandleConnection(conn, true)
	return nil
}
