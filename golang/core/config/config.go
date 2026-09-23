package config

type IConfig struct {
	Location ILocation
}

type ILocation struct {
	Download string // read location
	Store    string // patch the mongoose certification to validate the data
	Link     string // location to store the mongoose created links that can be used to push to the download string
}
