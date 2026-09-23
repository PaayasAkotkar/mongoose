package mongoose

import (
	msocket "app/mongoose/core/socket"
	"context"
	"log"
)

// Connect starts listening and handles each incoming connection/stream
// The server receives streams opened by clients
func (a *App) Connect(ctx context.Context) {
	addr := a.addr
	tconfig := a.tconfg
	qconfig := a.qconfig
	log.Println("udp server live on ", addr)

	if err := a.core.Listen(addr, tconfig, qconfig); err != nil {
		panic(err)
	}

	a.core.AsyncLIV(ctx, func(s *msocket.Socket) {
		select {
		case <-s.Conn().HandshakeComplete():
			a.updateState(StateConnected)
			sx := IState{
				stat: StateConnected,
				s:    s,
			}
			a.connected.Publish(CTKEY, &sx)

			log.Println("[server] handshake complete - client connected")
		case <-ctx.Done():
			a.updateState(StateClosed)
			return
		}

		for {
			select {
			case <-ctx.Done():
			default:
				m, err := s.ReadMessage()
				if err != nil {
					a.updateState(StateDisconnected)
					log.Printf("[server] stream closed: %s", err)
					return
				}
				if m.Ready {
					p := &IMessage{
						Message: m,
						s:       s,
						Error:   nil,
					}
					a.message.Publish(MKEY, p)
				} else {
					log.Printf("[server] message not ready")
				}
			}

			d := s.ReadDatagram(ctx)
			log.Println("data: ", d)
			if d.Ready {
				p := &IDMessage{
					Message: d,
					Error:   nil,
				}
				a.dmessage.Publish(DKEY, p)
			}
		}

	})
}

func (a *App) Stats(ctx context.Context, handler func(r *IDMessage)) {
	out := make(chan *IDMessage, 100)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result, ok := <-out:
				if !ok {
					return
				}
				if result != nil {
					handler(result)
				}
			}
		}
	}()

	go func() {
		go func(mid string) {
			ch := a.dmessage.Subscribe(mid)
			defer a.dmessage.Unsubscribe(mid, ch)

			for {
				select {
				case <-ctx.Done():
					return
				case result, ok := <-ch:
					if !ok {
						return
					}
					if result == nil {
						continue
					}
					select {
					case out <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}(DKEY)
	}()
}

// OnMessage registers a handler that is called for every message received
func (a *App) OnMessage(ctx context.Context, handler func(m *IMessage)) {
	out := make(chan *IMessage, 100)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result, ok := <-out:
				if !ok {
					return
				}
				if result != nil {
					handler(result)
				}
			}
		}
	}()

	go func() {
		go func(mid string) {
			ch := a.message.Subscribe(mid)
			defer a.message.Unsubscribe(mid, ch)

			for {
				select {
				case <-ctx.Done():
					return
				case result, ok := <-ch:
					if !ok {
						return
					}
					if result == nil {
						continue
					}
					select {
					case out <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}(MKEY)
	}()
}

// OnClose registers a handler called when the connection is fully closed
func (a *App) OnClose(ctx context.Context, handler func()) {
	go func() {
		ch := a.state.Subscribe(SKEY)
		defer a.state.Unsubscribe(SKEY, ch)
		for {
			select {
			case state, ok := <-ch:
				if !ok {
					return
				}
				if state != nil && *state == StateClosed {
					handler()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

type IState struct {
	stat ConnState
	s    *msocket.Socket
}

func (s *IState) Socket() *msocket.Socket {
	return s.s
}

// OnConnect registers a handler called when a client successfully connects
func (a *App) OnConnect(ctx context.Context, handler func(m *msocket.Socket)) {
	go func() {
		ch := a.connected.Subscribe(CTKEY)
		defer a.connected.Unsubscribe(CTKEY, ch)
		for {
			select {
			case state, ok := <-ch:
				if !ok {
					return
				}
				if state != nil && state.stat == StateConnected {
					handler(state.Socket())
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// OnDisconnect registers a handler called when a client disconnects
func (a *App) OnDisconnect(ctx context.Context, handler func()) {
	go func() {
		ch := a.state.Subscribe(SKEY)
		defer a.state.Unsubscribe(SKEY, ch)
		for {
			select {
			case state, ok := <-ch:
				if !ok {
					return
				}
				if state != nil && *state == StateDisconnected {
					handler()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// SendMessage writes a message on the shared server socket
func (a *App) SendMessage(payload []byte) error {
	if err := a.core.WriteMessage(payload); err != nil {
		return err
	}
	return nil
}

// Close gracefully closes the server and all connections
func (a *App) Close() error {
	a.updateState(StateClosed)
	if a.core != nil {
		return a.core.Close()
	}
	return nil
}
