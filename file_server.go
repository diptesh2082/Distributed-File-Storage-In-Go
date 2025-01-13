package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

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

type Payload struct {
	key  string
	data []byte
}

type Message struct {
	Payload any
}

func (s *Server) BroadcastData(p *Payload) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		// if err := peer.Send(p.data); err != nil {
		// 	log.Printf("Error sending data to peer %s: %s", addr, err)
		// 	delete(s.peers, addr)
		// }
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	fmt.Println("Broadcasting to", len(peers), "peers:", p)
	return gob.NewEncoder(mw).Encode(p)
}

func (s *Server) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	if err := s.store.Write(key, tee); err != nil {
		return err
	}

	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}
	p := &Payload{
		key:  key,
		data: buf.Bytes(),
	}
	// fmt.Println(buf.Bytes(), "www")
	// fmt.Println(p.data, "payload")
	return s.BroadcastData(p)
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
		case msg := <-s.Transport.Consume():
			// handle incoming messages from transport
			// msg := <-s.Transport.Consume()
			var p Payload
			fmt.Println("-------------------", msg.Payload)

			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p.data); err != nil {
				log.Printf("Error decoding payload: %s. Payload: %v", err, msg.Payload) // Log the error with payload
				log.Fatal(err)
			}
			fmt.Printf("Received message: %+v\n", p)
			// process message and store in store
			// ...
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
