package msocket

import "github.com/quic-go/quic-go"

func (s *Socket) getWriter() *quic.Stream {
	return s.stream
}

func (s *Socket) getReader() *quic.Stream {
	if s == nil || s.stream == nil {
		return nil
	}
	return s.stream
}
