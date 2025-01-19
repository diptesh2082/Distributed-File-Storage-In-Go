package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Default root folder name for storing files
const defaultRootFolderName = "Dipnetwork"

// CASPathTransFormFunc generates a PathKey from a given key by hashing it
func CASPathTransFormFunc(key string) PathKey {
	// Create a SHA-1 hash of the key
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	// Split the hash string into slices of a fixed block size
	blocksize := 5
	sliceslen := len(hashStr) / blocksize
	paths := make([]string, sliceslen)
	for i := 0; i < sliceslen; i++ {
		paths[i] = hashStr[(i * blocksize):((i + 1) * blocksize)]
	}

	// Return a PathKey with the path and filename
	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

// PathTransFormFunc is a function type that transforms a string into a PathKey
type PathTransFormFunc func(string) PathKey

// PathKey represents a structured path and filename
type PathKey struct {
	PathName string
	Filename string
}

// StoreOptes holds options for configuring a Store
type StoreOptes struct {
	PathTransFormFunc PathTransFormFunc
	Root              string
}

// DefaultPathTransFormFunc is a default transformation function that returns the key as both path and filename
var DefaultPathTransFormFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

// Store represents a storage system with configurable options
type Store struct {
	StoreOptes
}

// NewStore creates a new Store with the given options, applying defaults where necessary
func NewStore(opts StoreOptes) *Store {
	if opts.PathTransFormFunc == nil {
		opts.PathTransFormFunc = DefaultPathTransFormFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}

	return &Store{
		StoreOptes: opts,
	}
}

// FullPath returns the full path including the filename
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

// GetTheFirstPath returns the first segment of the path
func (p PathKey) GetTheFirstPath() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// Exists checks if a file with the given ID and key exists in the store
func (s *Store) Exists(ID, key string) bool {
	pathKey := s.PathTransFormFunc(key)
	pathAndFileName := s.Root + "/" + ID + "/" + pathKey.FullPath()
	fmt.Println("pathAndFileName exist ::: ", pathAndFileName)

	_, err := os.Stat(pathAndFileName)
	return !errors.Is(err, os.ErrNotExist)
}

// Clear removes all files in the root directory of the store
func (s *Store) Clear(key string) error {
	return os.RemoveAll(s.Root)
}

// Delete removes a file with the given ID and key from the store
func (s *Store) Delete(ID, key string) error {
	pathkey := s.PathTransFormFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathkey.Filename)
	}()
	pathAndFileName := s.Root + "/" + ID + "/" + pathkey.GetTheFirstPath()
	fmt.Println("pathAndFileName ::: ", pathAndFileName)
	log.Printf("Deleted file: %s", pathAndFileName)
	return os.RemoveAll(pathAndFileName)
}

// Read retrieves a file with the given ID and key from the store
func (s *Store) Read(ID, key string) (int64, io.Reader, error) {
	return s.readStream(ID, key)
}

// Write stores data from the reader into a file with the given ID and key
func (s *Store) Write(ID, key string, r io.Reader) (int64, error) {
	return s.writeStream(ID, key, r)
}

// readStream opens a file for reading and returns its size and a file handle
func (s *Store) readStream(ID, key string) (int64, *os.File, error) {
	PathKey := s.PathTransFormFunc(key)
	fullPathAndFileNameWithRoot := s.Root + "/" + ID + "/" + PathKey.FullPath()
	file, err := os.Open(fullPathAndFileNameWithRoot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}

// openFileForWriting creates a file for writing, ensuring the directory structure exists
func (s *Store) openFileForWriting(ID, key string) (*os.File, error) {
	pathkey := s.PathTransFormFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, ID, pathkey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	fullPathAndFileNameWithRoot := s.Root + "/" + ID + "/" + pathkey.FullPath()
	return os.Create(fullPathAndFileNameWithRoot)
}

// WriteDecrypt decrypts data from the reader and writes it to a file with the given ID and key
func (s *Store) WriteDecrypt(encKey []byte, ID, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(ID, key)
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

// writeStream writes data from the reader to a file with the given ID and key
func (s *Store) writeStream(ID, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(ID, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}
