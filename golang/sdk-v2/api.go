package mongoose

import (
	client "app/mongoose/client-sdk-v1"
	msocket "app/mongoose/core/socket"
	mcrypt "app/mongoose/crypt"
	"app/mongoose/misc"
	sdk "app/mongoose/sdk-v1"
	"app/mongoose/sdk-v2/admin"
	webaddress "app/mongoose/web-address"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/Deardrops/binpacker"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// GoMonitor returns the current file state report
func (a *App) GoMonitor(ctx context.Context, handler func(track *ITrackFile)) {
	out := make(chan *ITrackFile, 100)

	go func() {
		for {
			select {
			case res, ok := <-out:
				if !ok {
					return
				}
				if res == nil {
					continue
				}
				handler(res)
			case <-ctx.Done():
				return
			}
		}
	}()

	go func(key string) {
		ch := a.trackk.Subscribe(key)
		defer a.trackk.Unsubscribe(key, ch)
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-ch:
				if !ok {
					return
				}
				if r == nil {
					continue
				}
				select {
				case out <- r:
				case <-ctx.Done():
					return
				}
			}
		}
	}(TKey)
}

// Download monitors the incoming file chunk to be write on the os disk
// it auto writes to the bin disk
func (a *App) Download(ctx context.Context) {
	log.Println("[download...]")
	log.Println("my-id", a.myID)

	//	ch := a.startDownload.Subscribe(DKey)
	//	defer a.startDownload.Unsubscribe(DKey, ch)
	//
	//	go func() {
	//		for {
	//			select {
	//			case <-ctx.Done():
	//			case t := <-ch:
	//				if t != nil && t.proceed {
	//					a.runDownload(ctx)
	//				}
	//			}
	//		}
	//	}()

	a.recieve.OnConnect(ctx, func(m *msocket.Socket) {
		log.Println("connected")
	})

	a.recieve.OnClose(ctx, func() {
		log.Printf("udp-close %s", a.myID)
	})

	a.recieve.OnMessage(ctx, func(m *sdk.IMessage) {
		log.Println("rec packet | my-id", a.myID)
		su := &misc.SafeUnpacker{U: binpacker.NewUnpacker(binary.BigEndian, bytes.NewReader(m.Message.Payload))}
		//fnLen := su.Uint64()
		//fnBytes := su.Bytes(fnLen)
		//chunkIndex := su.Uint64()
		//chunkLen := su.Uint64()
		//chunk := su.Bytes(chunkLen)
		//fn := string(fnBytes)
		fnLen := su.Uint64()
		fnBytes := su.Bytes(fnLen)
		chunkIndex := su.Uint64()
		totalChunks := su.Uint64()
		chunkLen := su.Uint64()
		chunk := su.Bytes(chunkLen)
		fn := string(fnBytes)

		log.Println("chunkin : ", chunkIndex)
		log.Println("total: ", totalChunks)
		if su.Err != nil {
			log.Printf("[server unpack failed: %s]", su.Err)
			return
		}
		log.Println("peer data: ", a.track.PeerFileData)
		bin := a.config.Volume.Bin // save to the bin folder
		log.Println("bin: ", bin)
		//if w := misc.WriteChunkFile(fn, int(chunkIndex), chunk, bin); w.Err != nil {
		//	log.Println(w.Err)
		//	return
		//} else {
		//	a.track.SeedFileData.ChunkInfo.MarkExt = w.MarkExt
		//	a.track.SeedFileData.ChunkInfo.DirLocation = bin
		//}
		if w := WriteChunkFileV2(fn, chunk, bin); w.Err != nil {
			log.Println(w.Err)
			return
		} else {
			a.track.SeedFileData.ChunkInfo.MarkExt = w.MarkExt
			a.track.SeedFileData.ChunkInfo.DirLocation = bin
		}
		log.Println("writing done 🤗")
		r := m.Socket().Stat().BytesRead
		w := m.Socket().Stat().BytesWritten
		log.Println("stats read | write: ", r, w)
		a.updatePeerDownload(w, r)
		a.trackk.Publish(TKey, &ITrackFile{
			IsReady: true,
			Data:    a.track,
		})
	})
}

func WriteChunkFileV2(filename string, chunk []byte, src string) *misc.IWriteChunk {
	//mext := ".part_"
	w := &misc.IWriteChunk{
		//MarkExt: mext,
		Err: nil,
	}
	absSrc, err := filepath.Abs(src)
	if err != nil {
		w.Err = fmt.Errorf("failed to resolve src: %w", err)
		return w
	}

	if err := os.MkdirAll(absSrc, 0755); err != nil {
		w.Err = err
		return w
	}
	base := filepath.Base(filename)
	log.Println("base: ", base)
	log.Println("src: ", src)
	//ext := fmt.Sprintf("%s%s%d", base, mext, index) // join with this extension
	chunkPath := filepath.Join(absSrc, base)
	err = os.WriteFile(chunkPath, chunk, 0644)
	if err != nil {
		w.Err = err
		return w
	}
	return w
}

// GET watches for the download-req & download-complete-req
func (a *App) GET(ctx context.Context) {

	a.app.GET("/meta_info", func(c *gin.Context) {
		log.Println("[get...]")
		log.Println("my-id", a.myID)
		var sft SuceedFileTransfer
		var m mcrypt.MetaInfo
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
			return
		}

		errM := json.Unmarshal(body, &m)
		errSTF := json.Unmarshal(body, &sft)
		switch true {
		case errM == nil:
			f := &IPeerFileData{
				Self: TFileData{
					ChunkInfo: m.ChunkInfo,
					Custom:    m.CustomInfo,
				},
				Trackers: m.Config.URLs,
				FileInfo: m.FileInfo,
			}
			log.Println("meta info", m)
			go func() {
				switch true {
				case m.UDPMetaInfo:
					log.Println("id: ", a.myID)
					a.track.ToComplete = int(m.FileInfo.Size)
					a.track.setPeerFileData(f)
				case m.MetaInfo:
					log.Println("id: ", a.myID)
					// via tcp send the file tracking first
					for i, r := range m.Config.TCPs {
						w := webaddress.New(r)
						m.UDPMetaInfo = true
						m.MetaInfo = false
						// push the same thing to the server
						w.Request().Add(fmt.Sprintf("%d_data_pack%s", i, r), "GET", m.Pack()).Go()
					}
					a.track.ToComplete = int(m.FileInfo.Size)
					a.track.setPeerFileData(f)
					a.runDownload(ctx)
					a.startDownload.Publish(DKey, &IDownload{proceed: true})

				}
			}()
		case errSTF == nil:
			a.seed.Publish(SKey, &ISeed{
				true,
			})
		default:
			log.Println("recieved none")
		}

	})
}

// POST posts on file-done & triggers the local rust server
func (a *App) POST(ctx context.Context) {
	a.app.POST("/completion", func(c *gin.Context) {
		log.Println("[post...]")
		log.Println("my-id", a.myID)

		gi := a.gotInfo.Subscribe(IKey)
		se := a.seed.Subscribe(SKey)

		defer a.gotInfo.Unsubscribe(IKey, gi)
		defer a.seed.Unsubscribe(SKey, se)

		select {
		case t := <-gi:
			if t != nil && t.got {
				c.JSON(http.StatusOK, "file done")
				return
			}
			c.JSON(http.StatusInternalServerError, "transfer failed")

		case t := <-se:
			if t != nil && t.seed {
				// alert local rust to join the file chunk
				b := webaddress.New(a.base)
				lf := &LocalJoinFile{a.track.SeedFileData}
				b.Request().Add("seed_done", "GET", lf.Pack())
				c.JSON(http.StatusOK, "seed triggered")
				return
			}

		case <-ctx.Done():
			c.JSON(http.StatusServiceUnavailable, "shutting down")

		case <-c.Request.Context().Done():
			// client disconnected — don't leak the goroutine waiting forever
		}
	})
}

// runDownload reads the dir chunks and pushes
// note: the file-dir must contain the file chunk
// it runs around the config trackers + metainfo trackers if provided any
// todo: split the data into as per the servers
func (a *App) runDownload(ctx context.Context) {
	log.Println("my-id", a.myID)

	log.Println("running download")
	log.Println("peer-file-data", a.track.PeerFileData)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, track := range a.track.PeerFileData.Trackers {
			track := track
			log.Println("track", track)
			sender := client.New(track, a.config.TConf, a.config.QConf)
			sender.Connect(ctx, track)

			sender.OnConnect(ctx, func(m *msocket.Socket) {
				log.Println("sending packet....")
				dir := a.track.PeerFileData.Self.ChunkInfo.DirLocation
				wr := a.write(dir, m)
				g := &IGotInfo{
					false,
					SuceedFileTransfer{NumberOfFiles: wr.totalFile, SuceedTransfer: false},
				}
				if err := wr.err; err != nil {
					log.Println(err)
					//a.gotInfo.Publish(IKey, g)
					return
				}
				g.got = true
				// on suceed tell the current machine that the
				// chunks are successfully sent to the machine

				a.gotInfo.Publish(IKey, g)
				w := m.Stat().BytesWritten
				r := m.Stat().BytesRead

				log.Println("done writing packet")

				a.updatePeerStats(w, r)
			})
		}
	}()
	wg.Wait()
}

// Run runs the app without implicitly using the uber-fx package
// no need to handle the download, get, post api
// simply run this method
func (a *App) Run() {
	fx.New(
		fx.Provide(func() *sdk.App { return a.recieve }),
		fx.Provide(func() *gin.Engine { return a.app }),
		fx.Provide(func() admin.IApp {
			return admin.IApp{
				SendAddr:    a.serverAddrs.SendAddr,
				RecieveAddr: a.serverAddrs.RecieveAddr,
				Addr:        a.serverAddrs.Addr,
			}
		}),
		fx.Provide(func() admin.Handlers {
			return admin.Handlers{
				GET:      a.GET,
				POST:     a.POST,
				Download: a.Download,
			}
		}),
		fx.Invoke(admin.Start),
	).Run()

}
