package p2p

// import "net"

var (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
