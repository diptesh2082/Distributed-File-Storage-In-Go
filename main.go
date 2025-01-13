package main

import (
	"fmt"
	"log"
	"time"

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
		// OnPeer:        OnPeer,
	}
	// Create a new TCP transport instance
	TCPTransport := p2p.NewTCPTransport(tcpopts)
	fileServerOpts := ServerOptes{
		PathTransFormFunc: CASPathTransFormFunc,
		StorageRoot:       "4000_net",
		Transport:         TCPTransport,
	}
	fileServer := NewServer(fileServerOpts)
	go func ()  {
		time.Sleep(3 * time.Second)
		fileServer.Stop()
	}()

	err := fileServer.Start()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Error starting fileServer: %s\n", err)
		// return
	}

	// go func() {
	// 	for {
	// 		msg := <-tr.Consume()
	// 		fmt.Printf("%v\n", msg)
	// 	}
	// }()
	// // Start listening and accepting connections
	// err := tr.ListenAndAccept()
	// if err != nil {
	// 	log.Fatal(err)
	// 	fmt.Printf("Error starting transport: %s\n", err)
	// 	return
	// }

	// Keep the program running to accept incoming connections
	// select {}
}
