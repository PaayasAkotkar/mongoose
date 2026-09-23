package mstamp

import (
	"encoding/json"
	"log"
)

type ICerticate struct {
	Stamp IStamp `json:"stamp"`
	// end
	RProgress float64 `json:"read_progress"`  // current read-progress [related to gathering the info]
	WProgress float64 `json:"write_progress"` // current writing-progress [related to stamping on the os]
	// machine-to-handshake
	// end
	Succeed bool   `json:"succeed"` // the remote address written successfull
	Schema  Schema `json:"schema"`
}

func (c *ICerticate) Pack() []byte {
	p, err := json.Marshal(c)
	if err != nil {
		log.Println(err)
	}
	return p
}

type Schema struct {
	RemoteAddress IRemoteAddress `json:"remote_address"`
	Geo           Point          `json:"geo"`
}
type IRemoteAddress struct {
	Network Network `json:"network"`
	Owner   string  `json:"owner"`     // if-any
	OS      string  `json:"os"`        // machine's os name
	IP      string  `json:"remote_ip"` // machine's ip
}
type Network struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Peer         string `json:"peer"` // remote-address of the peer
	LocalAddress string `json:"local_address"`
	UDP          bool   `json:"udp"`
}

// Point implements the struct better for the h3 data
type Point struct {
	// Latitude & Longitude of the current machine
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// IStamp writes the mongoose stamp for the folder|file or any asset against malware protection
type IStamp struct {
	Policy IPolicy `json:"policy"`
	Stamp  bool    `json:"stamp"` // whether the stamping the machine is succeed
	ID     string  `json:"id"`    // hash-id
}
type IPolicy struct {
	Obey  bool   `json:"obey"`  // to obey that the folder is not a malware and if in-case malware the mongoose has rights to block the machine's drive
	Grant IGrant `json:"grant"` // grant permissions that to be accepted
}
type IGrant struct {
	Write bool `json:"write"` // to grant permimsion, write-down the validation verification files first so that process are not be repeated
}
