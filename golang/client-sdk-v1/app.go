package mongoose

import (
	cheetah "app/mongoose/cheetha"
	"app/mongoose/common"
	msocket "app/mongoose/core/socket"
	"app/mongoose/stack"
	"crypto/tls"
	"sync"

	"github.com/quic-go/quic-go"
)

type App struct {
	// core   *mongoose.IClient
	socket  *msocket.Socket // the live bidirectional stream socket
	State   ConnState
	mkeys   *stack.Stack[string] // keys related to monitor the message
	tconfig *tls.Config
	qconfig *quic.Config

	message   *cheetah.Cheetah[string, IMessage]
	connected *cheetah.Cheetah[string, IState]
	state     *cheetah.Cheetah[string, ConnState]
	isClosing chan bool
	mu        sync.Mutex
}

func New(addr string, tconfig *tls.Config, qconfig *quic.Config) *App {
	s := stack.NewStack[string]()
	s.Push(MKEY)

	a := &App{
		tconfig:   tconfig,
		qconfig:   qconfig,
		mkeys:     s,
		message:   cheetah.New[string, IMessage](100),
		connected: cheetah.New[string, IState](100),
		state:     cheetah.New[string, ConnState](100),
		isClosing: make(chan bool, 100),
		State:     StateClosed,
	}
	return a
}

type IState struct {
	stat ConnState
	s    *msocket.Socket
}

func (s *IState) Socket() *msocket.Socket {
	return s.s
}

const (
	MKEY  = "MESSAGE"
	CTKEY = "CONNECTED"
	SKEY  = "STATE"
)

const (
	CAP = 64 * common.KB
)

type ConnState int

const (
	StateDisconnected ConnState = iota // 0
	StateConnecting                    // 1
	StateConnected                     // 2
	StateClosed                        // 3
)

func (a *App) updateState(state ConnState) {
	a.mu.Lock()
	a.State = state
	a.mu.Unlock()
	a.state.Publish(SKEY, &state)
}

// setSocket stores the live socket so SendMessage can write to it.
func (a *App) setSocket(s *msocket.Socket) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.socket = s
}
func (a *App) Socket() *msocket.Socket {
	return a.socket
}

type IMessage struct {
	Message *msocket.Message
	s       *msocket.Socket
	Error   error
}

func (m *IMessage) SendMessage(payload []byte) error {
	if err := m.s.WriteMessage(payload); err != nil {
		return err
	}
	return nil
}
