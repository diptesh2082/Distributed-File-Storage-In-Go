package main

import (
	"fmt"
	"log"

	"github.com/diptesh/filestore/p2p"
)

func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListnerAdder:  ":4000",
		HandShakeFunc: p2p.NOPHandShake,
		Decoder:       &p2p.DefaultDecoder{},
	}
	// Create a new TCP transport instance
	tr := p2p.NewTCPTransport(tcpopts)

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
