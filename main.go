package main

import (
	"bytes"
	"fmt"
	"time"

	// "time"

	"github.com/diptesh/filestore/p2p"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	fmt.Println("New peer connected")
	return nil
}
func makeServer(ListnerAdder string, nodes ...string) *Server {
	TCPTransportOpts := p2p.TCPTransportOpts{
		ListnerAdder:  ListnerAdder,
		HandShakeFunc: p2p.NOPHandShake,
		Decoder:       &p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	// Create a new TCP transport instance
	TCPTransport := p2p.NewTCPTransport(TCPTransportOpts)
	fileServerOpts := ServerOptes{
		PathTransFormFunc: CASPathTransFormFunc,
		StorageRoot:       ListnerAdder + "_net",
		Transport:         TCPTransport,
		BootstrapNodes:    nodes,
	}
	s := NewServer(fileServerOpts)
	TCPTransport.OnPeer = s.OnPeer
	return s
}

func main() {

	fileServer1 := makeServer(":3000")
	fileServer2 := makeServer(":4000", ":3000")
	// go func ()  {
	// 	time.Sleep(3 * time.Second)
	// 	fileServer.Stop()
	// }()

	// go func() {
	// 	log.Fatal(fileServer1.Start())
	// }()
	// time.Sleep(1 * time.Second)
	// go func() {
	// 	log.Fatal(fileServer2.Start())
	// }()
	go fileServer1.Start()
	go fileServer2.Start()
	time.Sleep(1 * time.Second)
	data := bytes.NewReader([]byte("This is my  big file"))
	err := fileServer2.StoreData("mysecretdata", data)
	if err != nil {
		fmt.Println("error :: ", err)
	}
	select {}
}
