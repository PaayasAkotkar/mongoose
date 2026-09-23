package misc

import (
	"crypto/sha256"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

// PieceCount returns total number of pieces that may be cut
func PieceCount(totalSize, boundary int) float64 {
	return math.Ceil(float64(totalSize) / float64(boundary))
}

// CalculateHash returns the sha256 string
func CalculateHash(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}

// ReadSizeString implments the humanize package typially for reading the size
func ReadSizeString(b uint64) string {
	return humanize.Bytes(b)
}

// SplitIntoDivisions splits the ars into required div
func SplitIntoDivisions[Out any](chunks []Out, divisions int) (file [][]Out) {
	//if divisions < 1 {
	//	divisions = 1
	//}
	divisions = max(1, divisions)

	total := len(chunks)
	perDivision := int(math.Ceil(float64(total) / float64(divisions)))

	result := make([][]Out, 0, divisions)
	for i := 0; i < total; i += perDivision {
		end := min(i+perDivision, total)
		result = append(result, chunks[i:end])
	}
	copy(result, file)
	return
}

// AfterColon slices token if matched as per the regex
// the format i am working is X: data
// example:
// func usuage() {
//
//		type I struct {
//			Name string `json:"name"`
//		}
//
//		a := I{
//			Name: "jognny",
//		}
//		m, _ := json.Marshal(a)
//		p := "N: " + string(m)
//		var Conn = regexp.MustCompile(`N:\s*(.+)`)
//
//		x := AfterColon[[]byte](Conn, p)
//		var k I
//		json.Unmarshal(x, &k)
//		log.Println(k)
//		log.Println(x)
//	}
func AfterColon[T string | bool | int | []byte](rege *regexp.Regexp, token string) T {
	var zero T
	a := rege.FindStringSubmatch(token)

	switch any(zero).(type) {
	case int:
		n, _ := strconv.Atoi(strings.TrimSpace(a[1]))
		return any(n).(T)
	case bool:
		return any(strings.Contains(a[1], "true")).(T)
	case []byte:
		return any([]byte(a[1])).(T)
	default: // string
		return any(a[1]).(T)
	}
}

func SaveAs(dst, save string) string {
	return filepath.Join(dst, save)
}
