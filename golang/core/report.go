package mongoose

import (
	"app/mongoose/common"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"
)

type IReport struct {
	IsReady  bool    `json:"-"`
	Complete bool    `json:"-"`
	Monitor  Monitor `json:"monitor"`
}

type Monitor struct {
	FileName  string `json:"file_name"`
	Progress  string `json:"progress"`
	Ping      string `json:"ping"`
	Bandwidth string `json:"bandwidth"`
	FileSize  string `json:"file_size"`
	Peers     string `json:"peers"`
	Format    Time   `json:"time_format"`
	Status    Status `json:"status"`
}

type Time struct {
	TimeToComplete string `json:"time_to_complete"`
	TimeElapsed    string `json:"time_elapsed"`
}

type Status struct {
	Progress      string `json:"progress"`
	DownloadSpeed string `json:"download_speed"`
	UploadSpeed   string `json:"upload_speed"`
}

func (m *Monitor) ConvTimeToComplete(t time.Time) {
	m.Format.TimeToComplete = common.GenerateServerTime(t)
}

func (m *Monitor) ConvTimeToCompleteSecs(secs float64) {
	if math.IsInf(secs, 0) || math.IsNaN(secs) || secs < 0 {
		m.Format.TimeToComplete = "∞"
		return
	}
	d := time.Duration(secs) * time.Second
	if d < time.Minute {
		m.Format.TimeToComplete = fmt.Sprintf("%ds", int(d.Seconds()))
		return
	}
	if d < time.Hour {
		min := int(d.Minutes())
		sec := int(d.Seconds()) % 60
		if sec > 0 {
			m.Format.TimeToComplete = fmt.Sprintf("%dm %ds", min, sec)
		} else {
			m.Format.TimeToComplete = fmt.Sprintf("%dm", min)
		}
		return
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		min := int(d.Minutes()) % 60
		if min > 0 {
			m.Format.TimeToComplete = fmt.Sprintf("%dh %dm", h, min)
		} else {
			m.Format.TimeToComplete = fmt.Sprintf("%dh", h)
		}
		return
	}
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	if h > 0 {
		m.Format.TimeToComplete = fmt.Sprintf("%dd %dh", days, h)
	} else {
		m.Format.TimeToComplete = fmt.Sprintf("%dd", days)
	}
}

func (m *Monitor) ConvSize(size int64) {
	s := formatSize(float64(size))
	m.FileSize = fmt.Sprintf("%.2f%s", s.Size, s.Unit)
}

func (m *Monitor) ConvTimeElapsed(t time.Time) {
	m.Format.TimeElapsed = common.GenerateServerTime(t)
}

// ConvTimeElapsedSecs formats elapsed seconds as a human readable duration
func (m *Monitor) ConvTimeElapsedSecs(secs float64) {
	if math.IsInf(secs, 0) || math.IsNaN(secs) || secs < 0 {
		m.Format.TimeElapsed = "∞"
		return
	}
	d := time.Duration(secs) * time.Second
	if d < time.Minute {
		m.Format.TimeElapsed = fmt.Sprintf("%ds", int(d.Seconds()))
		return
	}
	if d < time.Hour {
		min := int(d.Minutes())
		sec := int(d.Seconds()) % 60
		if sec > 0 {
			m.Format.TimeElapsed = fmt.Sprintf("%dm %ds", min, sec)
		} else {
			m.Format.TimeElapsed = fmt.Sprintf("%dm", min)
		}
		return
	}

	if d < 24*time.Hour {
		h := int(d.Hours())
		min := int(d.Minutes()) % 60
		if min > 0 {
			m.Format.TimeElapsed = fmt.Sprintf("%dh %dm", h, min)
		} else {
			m.Format.TimeElapsed = fmt.Sprintf("%dh", h)
		}
		return
	}
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	if h > 0 {
		m.Format.TimeElapsed = fmt.Sprintf("%dd %dh", days, h)
	} else {
		m.Format.TimeElapsed = fmt.Sprintf("%dd", days)
	}
}

func (m *Monitor) ConvPogress(p float64) {
	m.Status.Progress = fmt.Sprintf("%.2f%%", p)
}

func speedStr(val float64) string {
	size := formatSize(val)
	return fmt.Sprintf("%.4f %s/s", size.Size, size.Unit)
}

func (m *Monitor) ConvSpeed(download, upload int64) {
	m.Status.DownloadSpeed = speedStr(float64(download))
	m.Status.UploadSpeed = speedStr(float64(upload))
}

func (m *Monitor) ConvPeer(value int) {
	m.Peers = fmt.Sprintf("%d", value)
}

func (m *Monitor) ConvPing(value string) {
	m.Ping = value
}

// end ~~~~~~~

func (m *Monitor) Pack() []byte {
	p, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil
	}
	return p
}
func (m *Monitor) GenerateReport() []byte {
	x, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil
	}
	return x
}
