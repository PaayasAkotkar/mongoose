package example

import (
	"app/mongoose/common"
	mcrypt "app/mongoose/crypt"
	"app/mongoose/misc"
	sdkv2 "app/mongoose/sdk-v2"
	webaddress "app/mongoose/web-address"
	"context"
	"log"
	"os"
	"sync"
	"time"
)

// SDKV2 showcase how the mongoose sdkv2 actaully
// transfer the file to multiple running server
func SDKV2() {

	ctx := context.Background()
	tc.InsecureSkipVerify = true

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		sdkv2_dummy_server_1(ctx)
	}()
	time.Sleep(3 * time.Second)
	go func() {
		defer wg.Done()
		sdkv2_dummy_client(ctx)
	}()
	time.Sleep(3 * time.Second)
	go func() {
		defer wg.Done()
		sdk_create_req(ctx)
	}()

	wg.Wait()
}

// sdkv2_dummy_server_1 recieves the rquest from client and writes down the file
func sdkv2_dummy_server_1(ctx context.Context) {

	cfg := &sdkv2.IConfig{
		TConf: tc,
		QConf: qc,
		Volume: sdkv2.IVolume{
			Bin: "./tests/bin/greet/new", // just for testing it can be any
		},
	}
	acfg := sdkv2.IApp{
		Addr:        "127.0.0.1:3000",
		SendAddr:    "127.0.0.1:3100",
		RecieveAddr: "127.0.0.1:3200",
	}

	var wg sync.WaitGroup

	app := sdkv2.New(&acfg, cfg, ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.GoMonitor(ctx, func(track *sdkv2.ITrackFile) {
			if track.IsReady {
				log.Println("doc: ", track.Data.Document())
				log.Println("data: ", track.Data.PeerFileData)
			}

		})
	}()

	//var wg sync.WaitGroup
	//wg.Add(3)
	//go func() {
	//	defer wg.Done()
	//	app.GET(ctx)
	//}()
	//go func() {
	//	defer wg.Done()
	//	app.POST(ctx)
	//}()
	//go func() {
	//	defer wg.Done()
	//	app.Download(ctx)
	//}()
	app.Run()
	wg.Wait()
}

func sdkv2_dummy_client(ctx context.Context) {

	cfg := &sdkv2.IConfig{
		TConf: tc,
		QConf: qc,
	}
	acfg := sdkv2.IApp{
		Addr:        "127.0.0.1:5000",
		SendAddr:    "127.0.0.1:5100",
		RecieveAddr: "127.0.0.1:5200",
	}
	app := sdkv2.New(&acfg, cfg, ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.GoMonitor(ctx, func(track *sdkv2.ITrackFile) {
			if track.IsReady {
				log.Println("doc: ", track.Data.Document())
				log.Println("data: ", track.Data.PeerFileData)

			}
		})
	}()

	//var wg sync.WaitGroup
	//wg.Add(3)
	//go func() {
	//	defer wg.Done()
	//	app.GET(ctx)
	//}()
	//go func() {
	//	defer wg.Done()
	//	app.POST(ctx)
	//}()
	//go func() {
	//	defer wg.Done()
	//	app.Download(ctx)
	//}()
	app.Run()
	wg.Wait()
}

func sdk_create_req(ctx context.Context) {
	m := &mcrypt.MetaInfo{}
	src := "./tests/samples/fisherchapbook87.pdf"

	bin := "./tests/bin/greet/aaa"

	li := misc.CreateLink(src, bin, m)
	log.Println("created link: ", li)

	m.CustomInfo.SaveFileNameAs = "suceed-result"
	m.ChunkInfo = mcrypt.IChunkInfo{
		DirLocation: bin,
		MarkExt:     "part_",
	}
	ma, err := os.Stat(src)
	if err != nil {
		panic(err)
	}
	x, err := os.ReadFile(src)
	if err != nil {
		panic(err)
	}
	m.FileInfo = mcrypt.IFileInfo{
		Name:      "greet.txt",
		Size:      uint64(ma.Size()),
		Extension: "txt",
		Hash:      misc.CalculateHash(string(x)),
	}
	m.CreatedDate = common.FormatDateForClient(time.Now())
	m.MetaInfo = true
	m.Config.URLs = []string{
		//"127.0.0.1:4200",
		"127.0.0.1:3200",
		//"127.0.0.1:3100",
	}
	m.Config.TCPs = []string{
		"http://127.0.0.1:3000/meta_info",
	}

	w := webaddress.New("http://127.0.0.1:5000")
	base := w.Path("meta_info").Generate()
	w.Request().SetBase(base)
	w.Request().Add("meta_info", "GET", m.Pack()).Go()
}
