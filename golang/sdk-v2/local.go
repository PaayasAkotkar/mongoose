package mongoose

import (
	"encoding/json"
	"log"
)

// todo in rust make sure
// to add the config for run localhost for rust file-joining

// LocalJoinFile message for rust to join the chunks
type LocalJoinFile struct {
	SeedFileData *ISeedFileData `json:"seed_file_data"`
	//
	//	DirLocation string `json:"dir_location"`
	//	SaveAs      string `json:"save_as"`
	//	MarkExt     string `json:"mark_ext"`
	//	Hash        string `json:"file_hash"`
}

func (j *LocalJoinFile) Pack() []byte {
	p, err := json.Marshal(j)
	if err != nil {
		log.Println(err)
		return nil
	}
	return p
}
