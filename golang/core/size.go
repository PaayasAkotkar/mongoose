package mongoose

import (
	"app/mongoose/common"
)

type ReadSize struct {
	Size float64
	Unit string
}

func formatSize(size float64) ReadSize {

	switch {
	case size >= common.TB:
		return ReadSize{Size: float64(size) / common.TB, Unit: "TB"}
	case size >= common.GB:
		return ReadSize{Size: float64(size) / common.GB, Unit: "GB"}
	case size >= common.MB:
		return ReadSize{Size: float64(size) / common.MB, Unit: "MB"}
	default:
		return ReadSize{Size: float64(size) / common.KB, Unit: "KB"}
	}

}
