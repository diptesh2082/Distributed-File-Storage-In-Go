package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"testing"
)

func TestCASPathTransFormFunc(t *testing.T) {
	key := "jsbfghjkrwbgfhjkersbgkwe"
	pathkey := CASPathTransFormFunc(key)
	expected := "4a29f/a1cd6/8dd1d/2f79e/905c0/77a6f/1a156/1af97"
	OriginalPathHash := "4a29fa1cd68dd1d2f79e905c077a6f1a1561af97"
	fmt.Printf("PathKey: %+v\n", pathkey)
	if pathkey.PathName != expected {
		t.Errorf("have %s got %s", expected, pathkey.PathName)
	}
	if pathkey.Filename != OriginalPathHash {
		t.Errorf("have %s got %s", OriginalPathHash, pathkey.Filename)
	}
}

func TestStoreDelete(t *testing.T) {
	opts := StoreOptes{
		PathTransFormFunc: CASPathTransFormFunc,
	}
	s := NewStore(opts)
	key := "mypicture"
	data := []byte("some image in byte")
	if err := s.WriteStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}

	exists, err := s.Exists(key)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Errorf("file %s should not exist after deletion", key)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOptes{
		PathTransFormFunc: CASPathTransFormFunc,
	}
	s := NewStore(opts)
	key := "mypicture"
	data := []byte("some image in byte")
	if err := s.WriteStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := ioutil.ReadAll(r)
	log.Printf("Read data: %s", b)
	if string(b) != string(data) {
		t.Errorf("have %s got %s", data, b)
	}
	// log.Printf("Read data: %s", b)
}
