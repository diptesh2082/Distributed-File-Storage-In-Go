package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents a peer in the TCP network.
type TCPPeer struct {
	net.Conn
	outbound bool
	wg       sync.WaitGroup
}

// TCPTransportOpts holds the options for configuring a TCPTransport.
type TCPTransportOpts struct {
	ListnerAddr   string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// TCPTransport manages the network transport layer for TCP connections.
type TCPTransport struct {
	TCPTransportOpts
	listner net.Listener
	mu      sync.RWMutex
	rpcch   chan RPC
}

// NewTCPPeer creates a new TCPPeer instance.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
	}
}

// NewTCPTransport creates a new TCPTransport instance.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		mu:               sync.RWMutex{},
		rpcch:            make(chan RPC, 1018),
	}
}

// Consume returns a channel to receive RPC messages.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close closes the TCPTransport listener.
func (p *TCPTransport) Close() error {
	return p.listner.Close()
}

// Send sends data over the TCP connection.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// ListenAndAccept starts listening for incoming connections and accepts them.
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listner, err = net.Listen("tcp", t.ListnerAddr)
	if err != nil {
		return err
	}

	fmt.Printf("Listening on %v\n", t.listner.Addr().String())
	go t.StartAcceptingLoop()
	return nil
}

// CloseStream signals that the stream is done.
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// Addr returns the address the transport is accepting connections on.
func (t *TCPTransport) Addr() string {
	return t.ListnerAddr
}

// StartAcceptingLoop continuously accepts incoming connections.
func (t *TCPTransport) StartAcceptingLoop() error {
	for {
		conn, err := t.listner.Accept()
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
			continue
		}
		go t.HandleConnection(conn, false)
	}
}

// HandleConnection manages a single connection, handling handshakes and message processing.
func (t *TCPTransport) HandleConnection(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("Dropping peer connection: %s\n", err)
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
		if err := t.Decoder.Decode(conn, rpc); err != nil {
			fmt.Printf("Error decoding message from %v: %s\n", peer.Conn.LocalAddr(), err)
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("Streaming to peer %s started\n", peer.Conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("Streaming ended to peer %s\n", peer.Conn.RemoteAddr())
			continue
		}

		t.rpcch <- *rpc
		fmt.Printf("Stream ended :: %s\n", conn.RemoteAddr())
	}
}

// Dial connects to a remote address and handles the connection.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.HandleConnection(conn, true)
	return nil
}
