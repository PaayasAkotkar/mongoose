// Package mcrypt ....
// meta-info.go writes the structure of the read write method more clearly
// making sure that the file is read via json first
package mcrypt

import (
	"encoding/json"
	"log"
)

type MetaInfo struct {
	// your-custom-fields
	ChunkInfo  IChunkInfo  `json:"chunk_info"`
	CustomInfo ICustomInfo `json:"custom_info"`
	Config     IConfig     `json:"config"`
	// end

	// can be ignored by the writer
	FileInfo    IFileInfo `json:"file_info"`
	CreatedDate string    `json:"created_date"`
	// end
	MetaInfo    bool `json:"meta_info"`     // came from the local-client
	UDPMetaInfo bool `json:"udp_meta_info"` // sent from the server itself to init the download tracker
}

type IFileInfo struct {
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	Extension string `json:"ext"`  // extension of the file attached
	Hash      string `json:"hash"` // a sha-256 todo: do not allow a file without a hash
}
type IChunkInfo struct {
	DirLocation string `json:"dir_location"` // location where the file is saved
	MarkExt     string `json:"mark_ext"`     // mark exention that is saved as part
	//Len uint64`json:"len"` // i am planning to use the field total number of chunks in the folder but again thought of mem usage
}
type ICustomInfo struct {
	SaveFileNameAs string   `json:"save_file_name_as"`
	Comment        IComment `json:"comment"`
}

type IConfig struct {
	URLs []string `json:"urls"` // data to fetch from ~ if empty the default trackers are being used
	TCPs []string `json:"tcps"` // related to non-udp urls -> set all the base value from the region to fetch
	Read bool     `json:"read"` // in-case of machine shutdown, on-read it writes from where it lefts
}

type IComment struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

func (m *MetaInfo) Pack() []byte {
	p, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil
	}
	return p
}
