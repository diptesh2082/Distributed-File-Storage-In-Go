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
	expectedFilePath := "4a29f/a1cd6/8dd1d/2f79e/905c0/77a6f/1a156/1af97"
	ExpectedFileName := "4a29fa1cd68dd1d2f79e905c077a6f1a1561af97"
	fmt.Printf("PathKey: %+v\n", pathkey)
	if pathkey.PathName != expectedFilePath {
		t.Errorf("have %s got %s", expectedFilePath, pathkey.PathName)
	}
	if pathkey.Filename != ExpectedFileName {
		t.Errorf("have %s got %s", ExpectedFileName, pathkey.Filename)
	}
}

// func TestStoreDelete(t *testing.T) {
// 	s := newStore()

// 	key := "mypicture"
// 	data := []byte("some image in byte")
// 	if err := s.Write(key, bytes.NewReader(data)); err != nil {
// 		t.Error(err)
// 	}

// 	if err := s.Delete(key); err != nil {
// 		t.Error(err)
// 	}

// 	exists, err := s.Exists(key)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if exists {
// 		t.Errorf("file %s should not exist after deletion", key)
// 	}
// }

func TestStore(t *testing.T) {
	s := newStore()
	ID := generateID()
	defer tearDown(t, s)
	for i := 0; i <= 50; i++ {
		key := "Dx" + fmt.Sprint(i)
		// ID := generateID()
		data := []byte("some image in byte")
		if _, err := s.writeStream(ID, key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		if ok := s.Exists(ID, key); !ok {
			t.Errorf("expected to have key %s", key)
		}
		n, r, err := s.Read(ID, key)
		if err != nil {
			t.Error(err)
		}

		b, _ := ioutil.ReadAll(r)
		log.Printf("Read data: %d", n)
		if string(b) != string(data) {
			t.Errorf("have %s got %s", data, b)
		}

		if err := s.Delete(ID, key); err != nil {
			t.Error(err)
		}

		if ok := s.Exists(ID, key); !ok {
			t.Errorf("expected to have key %s", key)
		}
	}
	// log.Printf("Read data: %s", b)
}

func newStore() *Store {
	opts := StoreOptes{
		PathTransFormFunc: CASPathTransFormFunc,
	}
	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store) {
	err := s.Clear(s.Root)
	if err != nil {
		t.Errorf("clear error: %s", err)
	}
}
