package p2p

import "net"

type Peer interface {
	// Close() error
	// RemoteAdder() net.Addr
	CloseStream()
	net.Conn
	Send([]byte) error
}

type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Dial(addr string) error
	// RemoteAdder() net.Addr
}
