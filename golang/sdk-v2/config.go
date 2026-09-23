package mongoose

import (
	"crypto/tls"

	"github.com/quic-go/quic-go"
)

type IConfig struct {
	Volume   IVolume
	Trackers ITrackers
	TConf    *tls.Config
	QConf    *quic.Config
	Ext      string // extension to add for the link
}

type IVolume struct {
	Store string // dir-location to store the results
	Link  string // dir-location to store the link
	Bin   string // dir-location of bin to store the file chunks
}

type track string

type ITrackers struct {
	trackers map[track][]string
	dTrack   track // default-tracker-family
}

// NewTracker inits the trackers
// make sure to set the default trackers else the seed seraching will be difficult
func NewTracker(backups []string) *ITrackers {
	t := make(map[track][]string)
	t[TBACKUP] = backups
	return &ITrackers{
		trackers: t,
	}
}

const (
	TPRIMARY   track = "PRIMARY"
	TSECONDARY track = "SECONDARY"
	TBACKUP    track = "BACKUP"
)

func (t *ITrackers) SetDefault(tracker track) {
	t.dTrack = tracker
}

func (t *ITrackers) SetTrackers(tracker track, addr string) {
	if t.trackers == nil {
		t.trackers = make(map[track][]string)
	}
	if tracker == TPRIMARY || tracker == TSECONDARY || tracker == TBACKUP {
		if _, ok := t.trackers[tracker]; ok {
			t.trackers[tracker] = append(t.trackers[tracker], addr)
		}
	}
}

// pullTrackers returns the seed address for that specific file
func (t *ITrackers) pullTrackers() []string {

	d := t.dTrack
	x := t.trackers[d]

	if x == nil {
		x = t.trackers[TBACKUP]
	}

	return x
}
