package main

import (
	"fmt"
	"log"

	"github.com/diptesh/filestore/p2p"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	fmt.Println("New peer connected")
	return nil
}
func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListnerAdder:  ":4000",
		HandShakeFunc: p2p.NOPHandShake,
		Decoder:       &p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	// Create a new TCP transport instance
	tr := p2p.NewTCPTransport(tcpopts)
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%v\n", msg)
		}
	}()
	// Start listening and accepting connections
	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Error starting transport: %s\n", err)
		return
	}

	// Keep the program running to accept incoming connections
	select {}
}
