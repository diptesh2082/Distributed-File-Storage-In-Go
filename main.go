package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/diptesh/filestore/p2p"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	fmt.Println("New peer connected")
	return nil
}
func makeServer(ListnerAdder string, nodes ...string) *Server {
	TCPTransportOpts := p2p.TCPTransportOpts{
		ListnerAddr:   ListnerAdder,
		HandShakeFunc: p2p.NOPHandShake,
		Decoder:       &p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	// Create a new TCP transport instance
	TCPTransport := p2p.NewTCPTransport(TCPTransportOpts)
	fileServerOpts := ServerOptes{
		EncKey:            newEncryptionKey(),
		PathTransFormFunc: CASPathTransFormFunc,
		StorageRoot:       ListnerAdder + "_net",
		Transport:         TCPTransport,
		BootstrapNodes:    nodes,
	}
	s := NewServer(fileServerOpts)
	TCPTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	// key := "mysecretdata"
	// key1 := "coolpicture.jpg"
	fileServer1 := makeServer(":3000")
	fileServer2 := makeServer(":4000", ":3000")
	fileServer3 := makeServer(":5000", ":3000", ":4000")
	// go func ()  {
	// 	time.Sleep(3 * time.Second)
	// 	fileServer.Stop()
	// }()

	go func() {
		log.Fatal(fileServer1.Start())
	}()
	// time.Sleep(1 * time.Second)
	// go func() {
	// 	log.Fatal(fileServer1.Start())
	// 	time.Sleep(1 * time.Second)
	// 	log.Fatal(fileServer2.Start())
	// }()
	// go fileServer1.Start()
	time.Sleep(1000 * time.Millisecond)
	go func() {
		log.Fatal(fileServer2.Start())
	}()
	time.Sleep(1000 * time.Millisecond)
	go fileServer3.Start()
	time.Sleep(1000 * time.Millisecond)

	// ******************** WRITE TEST *********************
	// data := bytes.NewReader([]byte("This is my picture file 12555"))
	// fmt.Println("Data byte size: ", data.Len())
	for i := 0; i < 10; i++ {
		key1 := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte(fmt.Sprintf("This is my  big file %s", fmt.Sprint(i))))
		err := fileServer3.StoreData(key1, data)
		if err != nil {
			fmt.Println("error :: ", err)
		}
		time.Sleep(100 * time.Millisecond)
		// //******************** DELETE TEST *********************
		err = fileServer3.store.Delete(fileServer3.ID, key1)
		if err != nil {
			log.Fatal(err)
		}
		//******************** READ TEST *********************
		r, err := fileServer3.GetData(key1)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))

	}
	// // //******************** DELETE TEST *********************
	// err := fileServer2.store.Delete(key1)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// //******************** READ TEST *********************
	// r, err := fileServer2.GetData(key1)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// b, err := ioutil.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(b))

	// select {}
}
