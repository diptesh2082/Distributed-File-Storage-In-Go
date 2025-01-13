package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func CASPathTransFormFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))

	// fmt.Printf("hash ::: %s\n", hash)

	hashStr := hex.EncodeToString(hash[:])

	// fmt.Printf("hashStr ::: %s\n", hashStr)

	blocksize := 5
	sliceslen := len(hashStr) / blocksize
	paths := make([]string, sliceslen)
	for i := 0; i < sliceslen; i++ {
		paths[i] = hashStr[(i * blocksize):((i + 1) * blocksize)]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTransFormFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}
type StoreOptes struct {
	PathTransFormFunc PathTransFormFunc
	Root              string
}

var DefaultPathTransFormFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOptes
}

func NewStore(opts StoreOptes) *Store {
	if opts.PathTransFormFunc == nil {
		opts.PathTransFormFunc = DefaultPathTransFormFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = "dipteshcomp"
	}
	return &Store{
		StoreOptes: opts,
	}
}
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

func (p PathKey) GetTheFirstPath() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (s *Store) Exists(key string) (bool, error) {
	pathkey := s.PathTransFormFunc(key)
	pathAndFileName := s.Root + "/" + pathkey.FullPath()
	_, err := os.Stat(pathAndFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransFormFunc(key)
	defer func() {
		log.Printf("Deleted file: %s", pathkey.Filename)
	}()
	pathAndFileName := s.Root + "/" + pathkey.GetTheFirstPath()
	log.Printf("Deleted file: %s", pathAndFileName)
	return os.RemoveAll(pathAndFileName)
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (*os.File, error) {
	PathKey := s.PathTransFormFunc(key)
	fullPathAndFileNameWithRoot := s.Root + "/" + PathKey.FullPath()
	return os.Open(fullPathAndFileNameWithRoot)
}

func (s *Store) WriteStream(key string, r io.Reader) error {
	pathkey := s.PathTransFormFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}
	// filename := "somefile"
	fullPathAndFileNameWithRoot := s.Root + "/" + pathkey.FullPath()

	f, err := os.Create(fullPathAndFileNameWithRoot)
	if err != nil {
		return nil
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to file %s", n, fullPathAndFileNameWithRoot)
	return nil
}
