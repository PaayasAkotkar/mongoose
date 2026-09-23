package msocket

import (
	"app/mongoose/common"
	mcrypt "app/mongoose/crypt"
	"app/mongoose/misc"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
)

type IStat struct {
	// BytesWritten current forward packet count
	BytesWritten int64
	// BytesRead current forward packet progress count
	BytesRead int64
}

const (
	ctrlWrite = math.MaxUint64 - 1
	ctrlRead  = math.MaxUint64 - 1
)

func (s *IStat) updateRead(c int64) {
	s.BytesRead += c
}
func (s *IStat) updateWrite(c int64) {
	s.BytesWritten += c
}

const (
	versionByte     = 1
	fixedHeaderSize = 1 + 1 + 4 // version + type + header
	maxHeaderSize   = 64 * common.KB
	maxPayloadSize  = 16 * common.MB
)

type Message struct {
	Payload []byte
	Ready   bool
}

// ParsePayload unmarshals the JSON payload into v
func (m *Message) ParsePayload(v any) error {
	if len(m.Payload) == 0 {
		return fmt.Errorf("[empty payload]")
	}
	return json.Unmarshal(m.Payload, v)
}

// RawPayload gives access to bytes directly, e.g. for file transfers (non-JSON)
func (m *Message) RawPayload() []byte {
	return m.Payload
}

type DMessage struct {
	Payload []byte
	Index   uint64
	Error   error
	Ready   bool
}

func (d *DMessage) Stats() *IStat {
	r := regexp.MustCompile(fmt.Sprintf(`%s\s*(.+)`, mcrypt.ReadKey))
	t := d.Payload
	if r.Match(t) {
		a := misc.AfterColon[int](r, string(t))
		return &IStat{
			BytesWritten: 0,
			BytesRead:    int64(a),
		}
	} else if r = regexp.MustCompile(fmt.Sprintf(`%s\s*(.+)`, mcrypt.WriteKey)); r.Match(t) {
		a := misc.AfterColon[int](r, string(t))
		return &IStat{
			BytesWritten: int64(a),
			BytesRead:    0,
		}
	}
	return &IStat{
		BytesWritten: 0,
		BytesRead:    0,
	}
}
