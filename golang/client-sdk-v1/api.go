package mongoose

import (
	msocket "app/mongoose/core/socket"
	"context"
	"fmt"
	"log"

	"github.com/quic-go/quic-go"
)

// Connect dials the server, opens a stream, and starts the continuous read loop
// The client opens the stream (server accepts it)
func (a *App) Connect(ctx context.Context, addr string) {
	go func() {
		a.updateState(StateConnecting)

		conn, err := quic.DialAddr(ctx, addr, a.tconfig, a.qconfig)
		if err != nil {
			log.Printf("[client] dial error: %s", err)
			a.updateState(StateDisconnected)
			return
		}
		defer conn.CloseWithError(0, "client done")

		select {
		case <-conn.HandshakeComplete():
			a.updateState(StateConnected)
			log.Println("[client] handshake complete – connected to server")
		case <-ctx.Done():
			a.updateState(StateClosed)
			return
		}

		stream, err := conn.OpenStreamSync(ctx)
		if err != nil {
			log.Printf("[client] open stream error: %s", err)
			a.updateState(StateDisconnected)
			return
		}
		defer stream.Close()

		s := msocket.New(conn, stream)
		a.setSocket(s)

		sx := IState{
			stat: StateConnected,
			s:    s,
		}
		a.connected.Publish(CTKEY, &sx)

		for {
			m, err := s.ReadMessage()
			if err != nil {
				a.updateState(StateDisconnected)
				log.Printf("[client] stream closed: %s", err)
				return
			}
			if m.Ready {
				p := &IMessage{
					Message: m,
					s:       s,
					Error:   nil,
				}
				a.message.Publish(MKEY, p)
			}
		}
	}()
}

// OnMessage registers a handler called for every message received from the server
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

// OnClose registers a handler called when the connection is closed
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

// OnConnect registers a handler called when the connection is established
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

// OnDisconnect registers a handler called when the client disconnects.
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

// SendMessage writes a framed message to the server over the open stream
func (a *App) SendMessage(payload []byte) error {
	//if len(payload) > CAP {
	//	return fmt.Errorf("package is too large to proceed")
	//}
	a.mu.Lock()
	s := a.socket
	a.mu.Unlock()
	if s == nil {
		return fmt.Errorf("[client] SendMessage called before connection is ready")
	}
	if err := s.WriteMessage(payload); err != nil {
		return fmt.Errorf("[client] SendMessage error: %s", err)
	}
	return nil
}

// Close gracefully closes the client connection
func (a *App) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.updateState(StateClosed)
	if a.socket != nil {
		if a.socket.Stream() != nil {
			a.socket.Stream().Close()
		}
		if a.socket.Conn() != nil {
			a.socket.Conn().CloseWithError(0, "client closed")
		}
	}
	return nil
}
