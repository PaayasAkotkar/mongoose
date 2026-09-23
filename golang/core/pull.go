package mongoose

type IStop struct {
	Stop     bool   `json:"stop"` // erases the stored cache of the packet
	FileName string `json:"file_name"`
}

type IDelete struct {
	Delete   bool   `json:"delete"` // erases whole packet
	FileName string `json:"file_name"`
}

type IPause struct {
	Pause    bool   `json:"pause"` // doesnt read the packet
	FileName string `json:"file_name"`
}
