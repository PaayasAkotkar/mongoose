package mongoose

import (
	cheetah "app/mongoose/cheetha"
	mongoose "app/mongoose/core"
	msocket "app/mongoose/core/socket"
	"app/mongoose/stack"
	"crypto/tls"
	"sync"

	"github.com/quic-go/quic-go"
)

type App struct {
	core *mongoose.App

	State ConnState
	mkeys *stack.Stack[string] // keys related to monitor the message

	addr      string
	tconfg    *tls.Config
	qconfig   *quic.Config
	connected *cheetah.Cheetah[string, IState]
	dmessage  *cheetah.Cheetah[string, IDMessage]
	message   *cheetah.Cheetah[string, IMessage]
	state     *cheetah.Cheetah[string, ConnState]
	isClosing chan bool
	mu        sync.Mutex
}

type IDMessage struct {
	Message *msocket.DMessage
	Error   error
}

func New(addr string, tconfig *tls.Config, qconfig *quic.Config) *App {
	s := stack.NewStack[string]()
	s.Push(MKEY)
	a := &App{
		core:      mongoose.New(),
		addr:      addr,
		tconfg:    tconfig,
		qconfig:   qconfig,
		mkeys:     s,
		dmessage:  cheetah.New[string, IDMessage](100),
		message:   cheetah.New[string, IMessage](100),
		connected: cheetah.New[string, IState](100),
		state:     cheetah.New[string, ConnState](100),
		isClosing: make(chan bool, 100),
		State:     3, // keep the conn close
	}
	//go a.Hub()

	return a
}

const (
	MKEY  = "MESSAGE"
	SKEY  = "STATE"
	CTKEY = "CONNECTED"
	DKEY  = "DATAGRAM-MESSAGE"
)

type ConnState int

const (
	StateDisconnected ConnState = iota // 0
	StateConnecting                    // 1
	StateConnected                     // 2
	StateClosed                        // 3
)

func (a *App) updateState(state ConnState) {
	a.state.Publish(SKEY, &state)
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

func (m *IMessage) Socket() *msocket.Socket {
	return m.s
}
