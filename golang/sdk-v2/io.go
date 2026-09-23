package mongoose

import (
	msocket "app/mongoose/core/socket"
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"path/filepath"

	"github.com/Deardrops/binpacker"
)

type IWrite struct {
	totalFile int
	err       error
}

// write pushes the file to the client
// it uses the rust created chunk-file to push it
//
//	func (a *App) write(dir string, s *msocket.Socket) *IWrite {
//		log.Printf("reading dir %s", dir)
//		a.mu.Lock()
//		defer a.mu.Unlock()
//
//		files, err := os.ReadDir(dir)
//		wx := &IWrite{
//			totalFile: -1,
//			err:       nil,
//		}
//		if err != nil {
//			wx.err = err
//			return wx
//		}
//
//		totalChunks := len(files)
//		wx.totalFile = totalChunks
//		log.Printf("[dir contains %d chunks] ", totalChunks)
//
//		for i, d := range files {
//			p := filepath.Join(dir, d.Name())
//
//			fs, err := os.Open(p)
//			if err != nil {
//				wx.err = err
//				return wx
//			}
//			defer fs.Close()
//			md, err := os.Stat(dir)
//			if err != nil {
//				wx.err = err
//				return wx
//			}
//			buf := make([]byte, md.Size())
//
//			log.Println("copying-dir", dir)
//			log.Println("copying-file", p)
//
//			for {
//				n, err := fs.Read(buf)
//				if n > 0 {
//					log.Println("pushing buf")
//
//					var _buf bytes.Buffer
//					fn := []byte(filepath.Base(p))
//					p := binpacker.NewPacker(binary.BigEndian, &_buf)
//
//					p.PushUint64(uint64(len(fn))).
//						PushBytes(fn).
//						PushUint64(uint64(i)).
//						PushUint64(uint64(totalChunks)).
//						PushUint64(uint64(len(buf[:n]))).
//						PushBytes(buf[:n])
//
//					log.Println("push done validating...")
//
//					if err := p.Error(); err != nil {
//						wx.err = err
//						return wx
//					}
//					log.Println("suceed push")
//					if err := s.WriteMessage(_buf.Bytes()); err != nil {
//						log.Println(err)
//					}
//					log.Println("suceeed write")
//				}
//				if err == io.EOF {
//					break
//				}
//				if err != nil {
//					log.Println(err)
//					wx.err = err
//					return wx
//				}
//				log.Println("read done")
//			}
//		}
//		wx.err = nil
//		return wx
//	}

// write pushes pre-chunked files from the directory using a memory-efficient stream loop
//func (a *App) write(dir string, s *msocket.Socket) *IWrite {
//	log.Printf("reading dir %s", dir)
//	a.mu.Lock()
//	defer a.mu.Unlock()
//
//	files, err := os.ReadDir(dir)
//	wx := &IWrite{
//		totalFile: -1,
//		err:       nil,
//	}
//	if err != nil {
//		wx.err = err
//		return wx
//	}
//
//	totalChunks := len(files)
//	wx.totalFile = totalChunks
//	log.Printf("[dir contains %d chunks] ", totalChunks)
//
//	for i, d := range files {
//		if d.IsDir() {
//			continue
//		}
//
//		p := filepath.Join(dir, d.Name())
//		fs, err := os.Open(p)
//		if err != nil {
//			wx.err = err
//			return wx
//		}
//
//		// Fixed small buffer (e.g., 10KB) to keep memory usage flat
//		buf := make([]byte, 10*common.KB)
//
//		log.Println("copying-file", p)
//
//		for {
//			n, err := fs.Read(buf)
//			if n > 0 {
//				var _buf bytes.Buffer
//				fn := []byte(d.Name())
//
//				// Pack with the exact slice length 'n' and data 'buf[:n]'
//				pk := binpacker.NewPacker(binary.BigEndian, &_buf)
//				pk.PushUint64(uint64(len(fn))).
//					PushBytes(fn).
//					PushUint64(uint64(i)).
//					PushUint64(uint64(totalChunks)).
//					PushUint64(uint64(n)).
//					PushBytes(buf[:n]) // Crucial: only exact bytes, no padding!
//
//				if err := pk.Error(); err != nil {
//					fs.Close()
//					wx.err = err
//					return wx
//				}
//
//				// Send via socket (WriteMessage handles the outer length prefix)
//				if err := s.WriteMessage(_buf.Bytes()); err != nil {
//					log.Println(err)
//					fs.Close()
//					wx.err = err
//					return wx
//				}
//
//				w := s.Stat().BytesWritten
//				r := s.Stat().BytesRead
//				a.updatePeerStats(w, r)
//
//				a.trackk.Publish(TKey, &ITrackFile{
//					IsReady: true,
//					Data:    a.track,
//				})
//			}
//
//			if err == io.EOF {
//				break
//			}
//			if err != nil {
//				log.Println(err)
//				fs.Close()
//				wx.err = err
//				return wx
//			}
//		}
//
//		fs.Close()
//	}
//	wx.err = nil
//	return wx
//}

// write pushes the file to the client
func (a *App) write(dir string, s *msocket.Socket) *IWrite {
	log.Printf("reading dir %s", dir)
	a.mu.Lock()
	defer a.mu.Unlock()

	files, err := os.ReadDir(dir)
	wx := &IWrite{
		totalFile: -1,
		err:       nil,
	}
	if err != nil {
		wx.err = err
		return wx
	}

	totalChunks := len(files)
	wx.totalFile = totalChunks
	log.Printf("[dir contains %d chunks] ", totalChunks)

	for i, d := range files {
		if d.IsDir() {
			continue // Skip subdirectories if any
		}

		p := filepath.Join(dir, d.Name())

		chunkData, err := os.ReadFile(p)

		if err != nil {
			wx.err = err
			return wx
		}

		if len(chunkData) == 0 {
			log.Println("skipping empty chunk file", p)
			continue
		}

		log.Println("copying-file-chunk", p)

		var _buf bytes.Buffer

		fn := []byte(filepath.Base(p))

		pk := binpacker.NewPacker(binary.BigEndian, &_buf)
		pk.PushUint64(uint64(len(fn))).
			PushBytes(fn).
			PushUint64(uint64(i)).
			PushUint64(uint64(totalChunks)).
			PushUint64(uint64(len(chunkData))).
			PushBytes(chunkData)

		if err := pk.Error(); err != nil {
			wx.err = err
			return wx
		}

		if err := s.WriteMessage(_buf.Bytes()); err != nil {
			log.Println(err)
			wx.err = err
			return wx
		}

		//
		//		w := s.Stat().BytesWritten
		//		r := s.Stat().BytesRead
		//		a.updatePeerStats(w, r)
		//
		//		a.trackk.Publish(TKey, &ITrackFile{
		//			IsReady: true,
		//			Data:    a.track,
		//		})
	}

	wx.err = nil
	return wx
}

// updateStats uses the byte-written as the source to update the stats
func (a *App) updatePeerStats(w, r int64) {

	a.track.updateWritten(int(w))
	a.track.updateProgress(w)
	a.track.updateBytesLeft(w)
	s := a.track.calcSpeed(r, w)
	a.track.updateTimeToComplete(int(s.Download))
}

// updatePeerDownload usese the byte-read as the priority to update the stats
func (a *App) updatePeerDownload(w, r int64) {

	a.track.updateWritten(int(r))
	a.track.updateProgress(r)
	a.track.updateBytesLeft(r)
	s := a.track.calcSpeed(r, w)
	a.track.updateTimeToComplete(int(s.Download))
}
