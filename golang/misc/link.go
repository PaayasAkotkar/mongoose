package misc

import (
	"app/mongoose/common"
	mcrypt "app/mongoose/crypt"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ICreateLink struct {
	ChunkInfo   mcrypt.IChunkInfo
	FileInfo    mcrypt.IFileInfo
	CreatedDate string
	Error       error
}

// CreateLink writes the link to the os to the config link storage & creates the file chunk
// that to be sent via golang
func CreateLink(fileLocation string, binLocation string, m *mcrypt.MetaInfo) *ICreateLink {
	l := &ICreateLink{}
	fs, err := os.Open(fileLocation)
	if err != nil {
		l.Error = err
		return l
	}
	defer fs.Close()

	st, err := fs.Stat()
	if err != nil {
		l.Error = err
		return l
	}
	m.FileInfo.Name = st.Name()
	m.FileInfo.Size = uint64(st.Size())
	m.FileInfo.Extension = filepath.Ext(fileLocation)
	hasher := md5.New()
	if _, err := io.Copy(hasher, fs); err != nil {
		l.Error = err
		return l
	}

	h := hex.EncodeToString(hasher.Sum(nil))
	m.FileInfo.Hash = h

	// seek back to beginning of file for chunk reading
	// because of hash
	if _, err := fs.Seek(0, io.SeekStart); err != nil {
		l.Error = err
		return l
	}

	log.Println("size: ", ReadSizeString(uint64(st.Size())))

	chunkSize := st.Size()
	buf := make([]byte, chunkSize) // i know this is horrible 💀 but now if its working we belive 😅

	var ace IChunk

	for {
		n, err := fs.Read(buf)
		log.Println("n: ", n)
		if n > 0 {
			log.Println("write")
			c := CreateChunk(buf[:n], int(chunkSize), 1, 8, 3)
			ace.Files = append(ace.Files, c.Files...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("nil", err)
			l.Error = err
			return l
		}
	}
	mext := ""
	id := strings.ReplaceAll(uuid.New().String(), "-", "_")
	px := filepath.Join(binLocation, id)
	for i, r := range ace.Files {
		if w := WriteChunkFile(m.FileInfo.Name, i, r.Chunk, px); w.Err != nil {
			l.Error = w.Err
			return l
		} else {
			mext = w.MarkExt
		}
	}
	l.ChunkInfo.DirLocation = px
	l.ChunkInfo.MarkExt = mext
	l.FileInfo = m.FileInfo
	l.CreatedDate = common.FormatDateForClient(time.Now())
	log.Println("meta: ", m)
	return l
}
