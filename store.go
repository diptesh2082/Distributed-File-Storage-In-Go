package main

import (
	// "bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "Dipnetwork"
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
	// ID                string
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
		opts.Root = defaultRootFolderName
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

func (s *Store) Exists(ID,key string) bool {
	pathKey := s.PathTransFormFunc(key)
	pathAndFileName := s.Root + "/" + ID + "/" + pathKey.FullPath()
	fmt.Println("pathAndFileName exist ::: ",pathAndFileName)

	_, err := os.Stat(pathAndFileName)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear(key string) error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(ID,key string) error {
	pathkey := s.PathTransFormFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathkey.Filename)
	}()
	pathAndFileName := s.Root + "/" + ID + "/" + pathkey.GetTheFirstPath()
	fmt.Println("pathAndFileName ::: ",pathAndFileName)
	log.Printf("Deleted file: %s", pathAndFileName)
	return os.RemoveAll(pathAndFileName)
}

func (s *Store) Read(ID,key string) (int64, io.Reader, error) {
	// n, f, err := s.readStream(key)
	// if err != nil {
	// 	return n, nil, err
	// }
	// defer f.Close()
	// buf := new(bytes.Buffer)
	// _, err = io.Copy(buf, f)

	return s.readStream(ID,key)
}
func (s *Store) Write(ID , key string, r io.Reader) (int64, error) {
	return s.writeStream(ID,key, r)
}

func (s *Store) readStream(ID,key string) (int64, *os.File, error) {
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
func (s *Store) openFileForWriting(ID,key string) (*os.File, error) {
	pathkey := s.PathTransFormFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, ID, pathkey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	// filename := "somefile"
	fullPathAndFileNameWithRoot := s.Root + "/" + ID + "/" + pathkey.FullPath()
	return os.Create(fullPathAndFileNameWithRoot)
}
func (s *Store) WriteDecrypt(encKey []byte, ID,key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(ID,key)
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Store) writeStream(ID,key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(ID,key)
	if err != nil {
		return 0, err
	}
	// defer f.Close()

	return io.Copy(f, r)
}
