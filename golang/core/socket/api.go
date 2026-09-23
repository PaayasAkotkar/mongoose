// Package msocket ...
// how things works
// uint32 => 4 bytes
// uint64 => 8 bytes
package msocket

import (
	metainfo "app/mongoose/core/meta-info"
	mcrypt "app/mongoose/crypt"
	merror "app/mongoose/errors"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Deardrops/binpacker"
	"github.com/quic-go/quic-go"
)

func (s *Socket) Conn() *quic.Conn {
	return s.conn
}

func (s *Socket) Stream() *quic.Stream {
	return s.stream
}

func (s *Socket) ReadPacket(location string) error {
	fs, err := os.ReadFile(location)
	if err != nil {
		return err
	}
	var mi metainfo.MetaInfo
	if err := json.Unmarshal(fs, &mi); err != nil {
		return err
	}
	return nil
}

// ReadMessage reads the message
// Simple framing: [payloadLen uint64][payload bytes]
func (s *Socket) ReadMessage() (*Message, error) {

	if s == nil {
		return nil, merror.NilError
	}

	r := s.getReader()

	if r == nil {
		return nil, fmt.Errorf("[uninitialized]")
	}

	b := binpacker.NewUnpacker(binary.BigEndian, r)
	payloadLen, err := b.ShiftUint64()
	if err != nil {
		return nil, err
	}
	payload, err := b.ShiftBytes(payloadLen)
	if err != nil {
		return nil, err
	}

	s.stat.updateRead(int64(len(payload)))
	s.read(int64(len(payload)))

	return &Message{
		Payload: payload,
		Ready:   true,
	}, nil
}

func (s *Socket) ReadDatagram(ctx context.Context) *DMessage {
	raw, err := s.conn.ReceiveDatagram(ctx)
	if err != nil {
		return &DMessage{
			Payload: raw,
			Index:   0,
			Ready:   true,
			Error:   err,
		}
	}

	// only 8 bytes are allowed
	if len(raw) < 8 {
		return &DMessage{
			Payload: raw,
			Index:   0,
			Ready:   true,
			Error:   fmt.Errorf("[datagram too short: %d bytes]", len(raw))}
	}

	index := binary.BigEndian.Uint64(raw[:8])
	data := raw[8:]

	return &DMessage{
		Payload: data, Index: index, Error: nil, Ready: true,
	}
}

// WriteMessage writes for the payload only
// Simple framing: [payloadLen uint64][payload bytes]
func (s *Socket) WriteMessage(payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r := s.getWriter()
	if r == nil {
		return fmt.Errorf("[write-message: uninitialized stream]")
	}
	b := binpacker.NewPacker(binary.BigEndian, r)
	b.PushUint64(uint64(len(payload))).
		PushBytes(payload)
	if err := b.Error(); err != nil {
		return fmt.Errorf("[write-message: %w]", err)
	}
	s.stat.updateWrite(int64(len(payload)))
	s.written(int64(len(payload)))
	return nil
}

// WriteDatagram send the data as the datagram
// frame: [payload len][payload-data]
func (s *Socket) WriteDatagram(data []byte, index uint64) error {
	buf := make([]byte, 8+len(data))
	binary.BigEndian.PutUint64(buf[:8], index)
	copy(buf[8:], data)
	// todo add the limit
	return s.conn.SendDatagram(buf)
}

func (s *Socket) written(count int64) {
	c := fmt.Sprintf("%s%d", mcrypt.WriteKey, count)
	log.Println("write: ", c)
	s.WriteDatagram([]byte(c), ctrlWrite)
}

func (s *Socket) read(count int64) {
	c := fmt.Sprintf("%s%d", mcrypt.ReadKey, count)
	log.Println("read: ", c)
	s.WriteDatagram([]byte(c), ctrlRead)
}

func (s *Socket) Stat() IStat {
	return s.stat
}
