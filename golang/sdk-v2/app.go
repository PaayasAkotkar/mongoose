package mongoose

import (
	cheetah "app/mongoose/cheetha"
	client "app/mongoose/client-sdk-v1"
	mcrypt "app/mongoose/crypt"
	sdk "app/mongoose/sdk-v1"
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
)

type App struct {
	base    string // url of the rust localhost
	myID    string // just for testing
	send    *client.App
	recieve *sdk.App

	config *IConfig

	track *TrackFile

	app         *gin.Engine
	serverAddrs *IApp

	startDownload *cheetah.Cheetah[string, IDownload]
	gotInfo       *cheetah.Cheetah[string, IGotInfo]
	seed          *cheetah.Cheetah[string, ISeed]
	create        *cheetah.Cheetah[string, ICreate]
	trackk        *cheetah.Cheetah[string, ITrackFile]
	wg            *sync.WaitGroup
	mu            sync.Mutex
}

type IApp struct {
	// udp address for quic
	SendAddr    string
	RecieveAddr string
	//end
	Addr string // tcp addr to run current app
}

func New(a *IApp, config *IConfig, ctx context.Context) *App {
	app := gin.New()
	tc, qc := config.TConf, config.QConf
	rec := sdk.New(a.RecieveAddr, tc, qc)
	sen := client.New(a.SendAddr, tc, qc)

	return &App{
		app:           app,
		myID:          fmt.Sprintf("tcp-%s |udp-send-%s |udp-rec-%s", a.Addr, a.SendAddr, a.RecieveAddr),
		recieve:       rec,
		send:          sen,
		config:        config,
		serverAddrs:   a,
		startDownload: cheetah.New[string, IDownload](100),
		gotInfo:       cheetah.New[string, IGotInfo](100),
		seed:          cheetah.New[string, ISeed](100),
		create:        cheetah.New[string, ICreate](1000),
		trackk:        cheetah.New[string, ITrackFile](100),
		track: &TrackFile{
			PeerFileData: &IPeerFileData{},
			SeedFileData: &ISeedFileData{},
			//IsComplete:   make(chan bool, 1)},
		}}
}

type ITrackFile struct {
	IsReady bool
	Data    *TrackFile
}

type ICreate struct {
	MetaInfo *mcrypt.MetaInfo
}

type ISeed struct {
	seed bool
}

type IDownload struct {
	proceed bool
}

type IGotInfo struct {
	got bool
	SuceedFileTransfer
}
