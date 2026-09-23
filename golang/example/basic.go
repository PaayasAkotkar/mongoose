package example

import (
	mongoose "app/mongoose/core"
	msocket "app/mongoose/core/socket"
	"context"
	"crypto/tls"
	"encoding/binary"
	"log"
	"sync"
	"time"

	"github.com/Deardrops/binpacker"
	"github.com/quic-go/quic-go"
)

const (
	addr = "127.0.0.1:3000"
)

var (
	tc = &tls.Config{
		Certificates: []tls.Certificate{mongoose.DefaultDevelopmentCertificate()},
		ServerName:   "localhost",
	}
	mon = mongoose.App{}
	qc  = &quic.Config{
		Allow0RTT:       true,
		EnableDatagrams: true,
	}
	wg sync.WaitGroup
)

func BASIC() {
	log.Println("[BASIC EXAMPLE]")
	//cert, err := mongoose.DefaultDevelopmentCertificate()
	//if err != nil {
	//	panic(err)
	//}

	ctx := context.Background()

	if err := mon.Listen(addr, tc, qc); err != nil {
		panic(err)
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		mon.AsyncLIV(ctx, func(s *msocket.Socket) {
			log.Println("[LIV CONNECTION]")
			for {
				m, err := s.ReadMessage()
				if err != nil {
					log.Printf("[client-disconnected] %s", err)
					return
				}
				iadd := s.Conn().RemoteAddr()
				log.Printf("peer %s", iadd.String())
				log.Printf("[package] %s", string(m.Payload))
			}
		})
	}()

	go func() {
		defer wg.Done()
		sampleRequest()
	}()
	wg.Wait()
}

func sampleRequest() {
	tc.InsecureSkipVerify = true
	ctx := context.Background()

	conn, err := quic.DialAddrEarly(ctx, addr, tc, qc)
	if err != nil {
		panic(err)
	}
	defer conn.CloseWithError(0, "client finished")

	select {
	case <-conn.HandshakeComplete():
	case <-ctx.Done():
		return
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		panic(err)
	}

	chunks := []byte("Hello Server")

	b := binpacker.NewPacker(binary.BigEndian, stream)
	b.PushUint64(1).PushUint64(uint64(len(chunks))).PushBytes(chunks)

	if err := b.Error(); err != nil {
		log.Println("push error:", err)
		return
	}

	if err := stream.Close(); err != nil {
		log.Println("stream close error:", err)
	}

	time.Sleep(500 * time.Millisecond)
	log.Println("request sent")
}
