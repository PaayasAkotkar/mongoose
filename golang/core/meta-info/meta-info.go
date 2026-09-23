package metainfo

import (
	"encoding/json"
	"log"
)

type MetaInfo struct {
	// your-custom-fields
	FileName string   `json:"file_name"`
	FileSize int64    `json:"file_size"`
	Comment  IComment `json:"comment"`
	URLs     []string `json:"urls"`
	// end

	Node        string    `json:"node"`
	Method      Method    `json:"method"`
	Hash        string    `json:"hash"`
	CreatedDate string    `json:"created_date"`
	FileInfo    IFileInfo `json:"file_info"`

	Read bool `json:"read"` // in-case of machine shutdown, on-read it writes from where it lefts
}

func ProgressLeft(total, currentSize int) float64 {
	if total == 0 {
		return 0
	}
	bytesLeft := float64(total - currentSize)
	return (bytesLeft / float64(total)) * 100
}

func (m *MetaInfo) Pack() []byte {
	p, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil
	}
	return p
}

type IComment struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

type Method struct {
	GET  bool `json:"get"`
	POST bool `json:"post"`
	LIV  bool `json:"liv"`
}

func (m *Method) Valid() bool {
	var g, p, l int
	if m.GET {
		g = 1
	}
	if m.POST {
		p = 1
	}
	if m.LIV {
		l = 1
	}

	return g+p+l == 1
}

type IMode struct {
	ReadOnly bool `json:"read_only"`
	Anyone   bool `json:"anyone"`
}

type IFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}
