package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listnerAdder := ":4000"
	tr := NewTCPTransport(listnerAdder)
	assert.Equal(t, tr.listnerAdder, listnerAdder)
	assert.Nil(t, tr.ListenAndAccept())
}
