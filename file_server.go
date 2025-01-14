package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/diptesh/filestore/p2p"
)

type ServerOptes struct {
	PathTransFormFunc PathTransFormFunc
	StorageRoot       string
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type Server struct {
	ServerOptes
	store    *Store
	quitech  chan struct{}
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
}

func NewServer(opts ServerOptes) *Server {
	// if opts.PathTransFormFunc == nil {
	// 	opts.PathTransFormFunc = DefaultPathTransFormFunc
	// }
	// if len(opts.Root) == 0 {
	// 	opts.Root = "dipteshcomp"
	// }

	storeOptes := StoreOptes{
		Root:              opts.StorageRoot,
		PathTransFormFunc: opts.PathTransFormFunc,
	}
	return &Server{
		ServerOptes: opts,
		store:       NewStore(storeOptes),
		quitech:     make(chan struct{}),
		peers:       make(map[string]p2p.Peer),
	}
}

// type DataMessage struct {
// 	key  string
// 	data []byte
// }

type Message struct {
	// From    string
	Payload any
}

func (s *Server) BroadcastData(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		// if err := peer.Send(p.data); err != nil {
		// 	log.Printf("Error sending data to peer %s: %s", addr, err)
		// 	delete(s.peers, addr)
		// }
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	fmt.Println("Broadcasting to", len(peers), "peers:", &msg)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *Server) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	msg := &Message{
		Payload: []byte("stroagekey"),
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		log.Printf("Error encoding message: %s", err)
		return err
	}
	for _, peer := range s.peers {
		// peers = append(peers, peer)
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}

	}
	time.Sleep(1 * time.Second)
	data := []byte("here is my  next big data")
	for _, peer := range s.peers {
		// peers = append(peers, peer)
		if err := peer.Send(data); err != nil {
			return err
		}

	}
	return nil
}

func (s *Server) loop() {
	defer func() {
		log.Printf("Server shutting down...\n")
		s.Transport.Close()
	}()
	for {
		select {
		case <-s.quitech:
			fmt.Println("-------------------")
			return
		case rpc := <-s.Transport.Consume():
			// handle incoming messages from transport
			// msg := <-s.Transport.Consume()
			var m Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				log.Printf("Error decoding payload: %s. Payload: %v", err, rpc.Payload) // Log the error with payload
				log.Fatal(err)
			}
			fmt.Printf("Received message: %s\n", string(m.Payload.([]byte)))
			// process message and store in store
			// ...
			peer, ok := s.peers[rpc.From]
			if !ok {
				panic("peer not found in peer map")
			}
			b := make([]byte, 10000)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}
			fmt.Printf("Received message for 2nd round: %s\n", string(m.Payload.([]byte)))

		}
	}
}

func (s *Server) OnPeer(p p2p.Peer) error {
	// function body
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("New peer connected from %s", p.RemoteAddr().String())

	return nil
}

func (s *Server) Start() error {
	log.Printf("Starting file server on %s...\n", s.StorageRoot)
	if err := s.Transport.ListenAndAccept(); err != nil {
		return fmt.Errorf("failed to start transport: %s", err)
	}
	if len(s.BootstrapNodes) != 0 {
		s.BootstrapNetwork()
	}
	s.loop()
	return nil
}

func (s *Server) Stop() {
	close(s.quitech)
}

func (s *Server) BootstrapNetwork() error {
	for _, adder := range s.BootstrapNodes {
		log.Printf("Tried to dial %s: %s\n", adder, adder)
		if len(adder) == 0 {
			continue
		}
		go func(adder string) {
			err := s.Transport.Dial(adder)
			if err != nil {
				log.Printf("Failed to dial %s: %s\n", adder, err)
				// return
			} else {
				log.Printf("Successfully dialed %s\n", adder)
			}
		}(adder)
	}
	return nil
}
