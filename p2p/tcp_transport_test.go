package p2p

import (
	"testing"

	// "github.com/diptesh/filestore/p2p"
	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tcpopts := TCPTransportOpts{
		ListnerAdder:  ":4000",
		HandShakeFunc: NOPHandShake,
		Decoder:       &DefaultDecoder{},
	}
	tr := NewTCPTransport(tcpopts)
	assert.Equal(t, tr.ListnerAdder, ":4000")
	assert.Nil(t, tr.ListenAndAccept())
}
