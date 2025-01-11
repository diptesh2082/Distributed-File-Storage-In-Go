package main

import (
	"fmt"
	"log"

	"github.com/diptesh/filestore/p2p"
)

func main() {
	// Create a new TCP transport instance
	tr := p2p.NewTCPTransport(":3000")

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
