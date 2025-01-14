package main

import (
	"bytes"
	"encoding/gob"
	"errors"
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

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
	// Size int64
}

type Message struct {
	// From    string
	Payload any
}

func (s *Server) Stream(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	// fmt.Println("Broadcasting to", len(peers), "peers:", &msg)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *Server) Broadcast(msg *Message) error {
	bufMsg := new(bytes.Buffer)
	if err := gob.NewEncoder(bufMsg).Encode(msg); err != nil {
		log.Printf("Error encoding message: %s", err)
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(bufMsg.Bytes()); err != nil {
			return err
		}

	}
	return nil
}

func (s *Server) GetData(key string) (io.Reader, error) {
	if !s.store.Exists(key) {
		return nil, errors.New("key does not exist")
		// log.Printf("Data  not avaible in the the disk")
	}
	r, err := s.store.Read(key)
	if err != nil {
		return nil, err
	}
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.Broadcast(&msg); err != nil {
		log.Printf("Error broadcasting message: %s", err)
		return nil, err
	}

	time.Sleep(1 * time.Second)
	log.Printf("Successfully Stroed bytes to Own Disk 1 ")

	for _, peer := range s.peers {
		FileBuffer := new(bytes.Buffer)
		n, err := io.Copy(FileBuffer, peer)
		if err != nil {
			log.Printf("Error copying data to peer: %s", err)
		}
		log.Printf("Successfully Stroed %d bytes to Own Disk", n)

	}
	log.Printf("Successfully Stroed bytes to Own Disk 2")

	select {}
	return r, nil
}

func (s *Server) StoreData(key string, r io.Reader) error {
	var (
		FileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, FileBuffer)
	)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.Broadcast(&msg); err != nil {
		log.Printf("Error encoding message: %s", err)
		return err
	}

	time.Sleep(1 * time.Second)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, FileBuffer)
		if err != nil {
			log.Printf("Error copying data to peer: %s", err)
		}
		log.Printf("Successfully Stroed %d bytes to Own Disk", n)

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
			return
		case rpc := <-s.Transport.Consume():

			var msg Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("Error decoding payload: %s. Payload: %v", err, rpc.Payload) // Log the error with payload
				// log.Fatal(err)
			}

			if err := s.HandleMessage(rpc.From, &msg); err != nil {
				log.Printf("Error handling message: %s", err)
			}
		}
	}
}

func (s *Server) HandleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.HandleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.HandleMessageGetFile(from, v)
	}
	return nil
}

func (s *Server) HandleMessageGetFile(from string, msg MessageGetFile) error {
	// fmt.Printf("Received message in HandleMessageGetFile : %v\n", (msg))
	if !s.store.Exists(msg.Key) {
		log.Printf("Ready to get the files from the disk")
		return errors.New("key does not exist")
	}
	r, err := s.store.Read(msg.Key)
	if err != nil {
		log.Printf("Error reading file: %s", err)
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in peer map", from)
	}
	n, err := io.Copy(peer, r)
	if err != nil {
		log.Printf("Error copying data to peer: %s", err)
		return nil
	}
	log.Printf("Successfully sent %d bytes to peer", n)
	return nil
}

func (s *Server) HandleMessageStoreFile(from string, msg MessageStoreFile) error {
	// fmt.Printf("Received message in HandleMessageStoreFile : %v\n", (msg))
	peer, ok := s.peers[from]
	if !ok {
		panic("peer not found in peer map")
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		// panic(err)
		return err
	}
	log.Printf("Successfully Stroed %d bytes to peer", n)
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
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

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
