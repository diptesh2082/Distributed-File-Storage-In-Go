package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/diptesh/filestore/p2p"
)

// ServerOptes contains configuration options for the file server
type ServerOptes struct {
	ID                string
	PathTransFormFunc PathTransFormFunc // Function to transform file paths
	StorageRoot       string            // Root directory for file storage
	Transport         p2p.Transport     // Network transport layer
	BootstrapNodes    []string          // List of bootstrap node addresses
	EncKey            []byte            // Encryption key for file encryption/decryption
}

// Server represents the main file server instance
type Server struct {
	ServerOptes
	store    *Store              // File storage backend
	quitech  chan struct{}       // Channel for shutdown signaling
	peerLock sync.Mutex          // Mutex for peer map access
	peers    map[string]p2p.Peer // Map of connected peers
}

// NewServer creates and initializes a new Server instance
func NewServer(opts ServerOptes) *Server {
	storeOptes := StoreOptes{
		Root:              opts.StorageRoot,
		PathTransFormFunc: opts.PathTransFormFunc,
	}
	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}
	return &Server{
		ServerOptes: opts,
		store:       NewStore(storeOptes),
		quitech:     make(chan struct{}),
		peers:       make(map[string]p2p.Peer),
	}
}

// MessageStoreFile represents a file storage message
type MessageStoreFile struct {
	ID string
	Key  string
	Size int64
}

// MessageGetFile represents a file retrieval message
type MessageGetFile struct {
	ID string
	Key string
}

// Message represents a generic message with a payload
type Message struct {
	Payload any
}

// Stream sends a message to all connected peers using a MultiWriter
func (s *Server) Stream(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

// Broadcast sends a message to all connected peers individually
func (s *Server) Broadcast(msg *Message) error {
	bufMsg := new(bytes.Buffer)
	log.Printf("Incoming encoding message: %v", msg.Payload)

	if err := gob.NewEncoder(bufMsg).Encode(msg); err != nil {
		log.Printf("Error encoding message: %v", err)
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(bufMsg.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// GetData retrieves file data either from local storage or from the network
func (s *Server) GetData(key string) (io.Reader, error) {
	if s.store.Exists(s.ID, key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}
	fmt.Printf("[%s] dont have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			ID: s.ID,
			Key: hashKey(key),
		},
	}
	if err := s.Broadcast(&msg); err != nil {
		log.Printf("Error broadcasting message: %s", err)
		return nil, err
	}
	log.Printf("Error encoding message: %s\n", msg.Payload)

	time.Sleep(100 * time.Millisecond)

	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		log.Printf("Successfully Stored %d bytes to Own Disk 1 ", fileSize)

		n, err := s.store.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] received (%d) bytes over the network from (%s)", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}
	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

// StoreData stores file data locally and broadcasts it to peers
func (s *Server) StoreData(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)
	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to file %s", size, s.StorageRoot)
	msg := Message{
		Payload: MessageStoreFile{
			ID: s.ID,
			Key:  hashKey(key),
			Size: size + 16,
		},
	}

	if err := s.Broadcast(&msg); err != nil {
		log.Printf("Error encoding message: %v", err)
		return err
	}

	time.Sleep(100 * time.Millisecond)
	log.Printf("encoding message 2 : %v", msg.Payload)

	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

	return nil
}

// loop runs the main server event loop
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
				log.Printf("Error decoding payload: %s. Message: %v", err, rpc)
			}
			if err := s.HandleMessage(rpc.From, &msg); err != nil {
				log.Printf("Error handling message: %s", err)
			}
		}
	}
}

// HandleMessage routes messages to appropriate handlers based on payload type
func (s *Server) HandleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.HandleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.HandleMessageGetFile(from, v)
	}
	return nil
}

// HandleMessageGetFile handles file retrieval requests from peers
func (s *Server) HandleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Exists(msg.ID, msg.Key) {
		log.Printf("Ready to get the files from the disk")
		return errors.New("key does not exist")
	}
	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		log.Printf("Error reading file: %s", err)
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Printf("closing readCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in peer map", from)
	}
	peer.Send([]byte{p2p.IncomingStream})

	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		log.Printf("Error copying data to peer: %s", err)
		return err
	}
	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)
	return nil
}

// HandleMessageStoreFile handles file storage requests from peers
func (s *Server) HandleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		panic("peer not found in peer map")
	}
	n, err := s.store.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	log.Printf("Successfully Stored %d bytes to peer", n)
	peer.CloseStream()
	return nil
}

// OnPeer handles new peer connections
func (s *Server) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("New peer connected from %s", p.RemoteAddr().String())
	return nil
}

// Start initializes and starts the file server
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

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	close(s.quitech)
}

// BootstrapNetwork connects to initial bootstrap nodes
func (s *Server) BootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		log.Printf("Tried to dial %s: %s\n", addr, addr)
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			err := s.Transport.Dial(addr)
			if err != nil {
				log.Printf("Failed to dial %s: %s\n", addr, err)
			} else {
				log.Printf("Successfully dialed %s\n", addr)
			}
		}(addr)
	}
	return nil
}

// init registers message types with gob encoder
func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
