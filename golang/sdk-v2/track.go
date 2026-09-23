package mongoose

import (
	mcrypt "app/mongoose/crypt"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

type TrackFile struct {
	ToComplete     int // constant file-size
	PeerFileData   *IPeerFileData
	SeedFileData   *ISeedFileData
	Written        int     // bytes written so far
	Progress       float64 // (written/ToComplete)*100
	TimeToComplete float64 // (ToComplete-bytes-left)
	ETA            float64 // todo
	//IsComplete          chan bool
	BytesLeft           int64 // to_complete-bytes-writte
	lastWrite, lastRead int64
	lastTime            time.Time // last time when the peer & seed communicate
	mu                  sync.Mutex
}

// update
func (t *TrackFile) setToComplete(size int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ToComplete = size
}

func (t *TrackFile) updateBytesLeft(w int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.BytesLeft = int64(t.ToComplete) - w
	log.Println("uP: ", t.BytesLeft)
}

func (t *TrackFile) setPeerFileData(f *IPeerFileData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.PeerFileData = f
}

func (t *TrackFile) setSeedFileData(f *ISeedFileData) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.SeedFileData = f
}

func (t *TrackFile) updateWritten(c int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Written = c
}

func (t *TrackFile) updateProgress(bytesWritten int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	log.Println("x: ", bytesWritten, t.ToComplete)
	log.Println("p: ", (bytesWritten/int64(t.ToComplete))*100)
	if t.ToComplete <= 0 {
		t.Progress = 0
		return
	}
	p := (float64(bytesWritten) / float64(t.ToComplete)) * 100
	//if p == 100.0 {
	//	t.IsComplete <- true
	//}
	t.Progress = p
}

// updateTimeToComplete the downloadSpeedBytesPerSec typically needs the download-speed
// use the calcSpeed to get the param for download speed
func (t *TrackFile) updateTimeToComplete(downloadSpeedBytesPerSec int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	bl := t.BytesLeft
	rem := t.ToComplete - int(bl)
	if rem <= 0 {
		t.TimeToComplete = 0
	}
	if downloadSpeedBytesPerSec <= 0 {
		t.TimeToComplete = math.Inf(1)
	}
	t.TimeToComplete = float64(rem) / float64(downloadSpeedBytesPerSec)

}

// end

type ISpeed struct {
	Download int64
	Upload   int64
}

// calcSpeed 😅 i forogt the pdf cause its from my old code
// i just modified it using the max func of go
func (t *TrackFile) calcSpeed(bytesRead int64, bytesWritten int64) *ISpeed {
	// old code
	// m.mu.Lock()
	//	defer m.mu.Unlock()
	//
	//	conn := t.Stats().AllConnStats
	//	read := conn.BytesRead.Int64()
	//	write := conn.BytesWritten.Int64()
	//
	//	var dSpeed, uSpeed int64
	//	if m.lastRead > 0 {
	//		dSpeed = read - m.lastRead
	//	} else if read > 0 {
	//		dSpeed = read
	//	}
	//	if m.lastWrite > 0 {
	//		uSpeed = write - m.lastWrite
	//	} else if write > 0 {
	//		uSpeed = write
	//	}
	//
	//	m.lastRead = read
	//	m.lastWrite = write
	//
	//	return &ISpeed{Download: dSpeed, Upload: uSpeed}
	// end
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()

	var dSpeed, uSpeed int64
	elapsed := now.Sub(t.lastTime).Seconds()
	if !t.lastTime.IsZero() && elapsed > 0 {
		dSpeed = int64(float64(max(bytesRead-t.lastRead, 0)) / elapsed)
		uSpeed = int64(float64(max(bytesWritten-t.lastWrite, 0)) / elapsed)
	}

	t.lastRead = bytesRead
	t.lastWrite = bytesWritten
	t.lastTime = now

	return &ISpeed{Download: dSpeed, Upload: uSpeed}
}

// Document a monitor report that can be read
// note: it not a json | yaml | bencode formart
func (t *TrackFile) Document() string {
	tc := fmt.Sprintf(`%d`, t.ToComplete)
	fn := t.PeerFileData.FileInfo.Name
	wr := fmt.Sprintf(`%d`, t.Written)
	prg := fmt.Sprintf(`%.4f`, t.Progress)
	toc := fmt.Sprintf(`%.4f`, t.TimeToComplete)
	eta := fmt.Sprintf(`%.4f`, t.ETA)

	return fmt.Sprintf(` 
	------ FILE-TRACK-REPORT ------
	to-complete:   %s
	file-name:     %s
	file-written:  %s
	progress-done: %s
	time-to-complete: %s
	ETA: %ss
	`, tc, fn, wr, prg, toc, eta)
}

type IPeerFileData struct {
	FileInfo mcrypt.IFileInfo
	Self     TFileData
	Trackers []string // trackers provided by the current file if any
}
type ISeedFileData struct {
	ChunkInfo mcrypt.IChunkInfo
}
type TFileData struct {
	ChunkInfo mcrypt.IChunkInfo
	Custom    mcrypt.ICustomInfo

	//FileName    string // actual file name
	//SaveAs      string // name of the after join-complete
	//DirLocation string // of file chunks to join
	//Ext         string // extension of the file
	//MarkExt     string // mark exention that is saved as part
	//FileHash    string // to-track file read-write process
}
