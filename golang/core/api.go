package mongoose

import (
	msocket "app/mongoose/core/socket"
	merror "app/mongoose/errors"
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/quic-go/quic-go"
)

// Listen starts the app engine
func (a *App) Listen(addr string, tcfg *tls.Config, cfg *quic.Config) error {
	li, err := quic.ListenAddr(addr, tcfg, cfg)
	if err != nil {
		return fmt.Errorf("[listen error %s]", err.Error())
	}
	if li == nil {
		return merror.NilError
	}
	a.lis = li
	return nil
}

func (a *App) WriteMessage(payload []byte) error {
	if a.msocket == nil {
		return nil
	}
	if err := a.msocket.WriteMessage(payload); err != nil {
		return err
	}
	return nil
}

// Close gracefully closes the listener and all connections
func (a *App) Close() error {
	if a.lis != nil {
		if err := a.lis.Close(); err != nil {
			return err
		}
	}
	if a.msocket != nil {
		if a.msocket.Stream() != nil {
			a.msocket.Stream().Close()
		}
		if a.msocket.Conn() != nil {
			a.msocket.Conn().CloseWithError(0, "server closed")
		}
	}
	return nil
}

// AsyncLIV handles long-lived, persistent packet streaming loops related to the stream
func (a *App) AsyncLIV(ctx context.Context, handler func(s *msocket.Socket)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := a.lis.Accept(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[accept-error] %s", err)
				continue
			}

			go func(c *quic.Conn) {
				defer c.CloseWithError(0, "connection closed by server loop")

				for {
					str, err := c.AcceptStream(ctx)
					if err != nil {
						return
					}
					go func(stream *quic.Stream) {
						defer stream.Close()
						st := msocket.New(c, stream)
						a.msocket = st
						handler(st)
					}(str)
				}
			}(conn)
		}
	}
}
