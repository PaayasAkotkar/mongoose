package mongoose

// SuceedFileTransfer implements all the udp packets are sent
// suceessfully
type SuceedFileTransfer struct {
	NumberOfFiles  int  `json:"number_of_files"`
	SuceedTransfer bool `json:"suceed_transfer"`
}
