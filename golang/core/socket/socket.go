package msocket

import (
	merror "app/mongoose/errors"
	"fmt"
	"sync"

	"github.com/quic-go/quic-go"
)

type Socket struct {
	stream *quic.Stream
	conn   *quic.Conn
	mu     sync.Mutex
	stat   IStat
}

func New(conn *quic.Conn, stream *quic.Stream) *Socket {
	if conn == nil || stream == nil {
		fmt.Println(merror.NilError)
		return nil
	}
	return &Socket{
		stream: stream,
		conn:   conn,
	}
}
