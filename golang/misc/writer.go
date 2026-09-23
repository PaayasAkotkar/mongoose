package misc

import (
	"fmt"

	"github.com/Deardrops/binpacker"
)

type SafeUnpacker struct {
	U   *binpacker.Unpacker
	Err error
}

func (s *SafeUnpacker) Uint64() uint64 {
	if s.Err != nil {
		return 0
	}
	v, err := s.U.ShiftUint64()
	if err != nil {
		s.Err = fmt.Errorf("shift uint64: %w", err)
	}
	return v
}

func (s *SafeUnpacker) Bytes(n uint64) []byte {
	if s.Err != nil {
		return nil
	}
	v, err := s.U.ShiftBytes(n)
	if err != nil {
		s.Err = fmt.Errorf("shift bytes(%d): %w", n, err)
	}
	return v
}
