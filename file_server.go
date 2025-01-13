package main

import (
	"fmt"
	"log"

	"github.com/diptesh/filestore/p2p"
)

type ServerOptes struct {
	PathTransFormFunc PathTransFormFunc
	StorageRoot       string
	Transport         p2p.Transport
}

type Server struct {
	ServerOptes
	store   *Store
	quitech chan struct{}
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
	}
}

func (s *Server) Start() error {
	log.Printf("Starting file server on %s...\n", s.StorageRoot)
	if err := s.Transport.ListenAndAccept(); err != nil {
		return fmt.Errorf("failed to start transport: %s", err)
	}
	s.loop()
	return nil
}

func (s *Server) Stop() {
	close(s.quitech)
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
			fmt.Printf("Received message: %v\n", msg)
			// process message and store in store
			// ...
		}
	}
}
