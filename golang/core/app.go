package mongoose

import (
	cheetah "app/mongoose/cheetha"
	"app/mongoose/core/config"
	msocket "app/mongoose/core/socket"

	"github.com/quic-go/quic-go"
)

type App struct {
	lis     *quic.Listener
	Config  *config.IConfig
	msocket *msocket.Socket
	// this typically acts like a direcrtory
	// you pick up the tracker and simply pull the file from there
	// less trackers are beneficial for searching mec
	Trackers []string // note: more trackers are generally not good for single file
	// end
	cheetah *cheetah.Cheetah[string, IReport]
}

func New() *App {
	return &App{
		cheetah: cheetah.New[string, IReport](100)}
}

// Download from progress
// in-case of machine failure => read the file and go to that page and simply start writing from that page onwords
