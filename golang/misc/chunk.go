package misc

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

// IFile implements the file data of the chunk file
type IFile struct {
	Chunk []byte `json:"chunk"`
	Range IRange `json:"range"` // current file range
	Stamp string `json:"stamp"` // sha-256
}

// IRange can consider it as how big is the chunk range like to how one chunk
type IRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// IChunk implements the strct that delivers the dataset of the chunk file
type IChunk struct {
	Files []IFile
	Split [][]IFile
}

// AEParams represents the required param for the AEC
type AEParams struct {
	TargetSize  int // expected chunk size
	Window      int // w — passed as lookAhead to AEChunking
	MaxBoundary int // maxBoundary — passed as maxBoundary to AEChunking
}

// FindSizes returns the target for the computeAEP
// fileSize len(packet)
// chunk how much splits are you planning to do for current file
// cap the max limit for internet udp packet is 32KB
// cap can be any but not corssing the limit of 32KB
// for perfect size you can even send any higher value to test
// don't try to send less value for higher files
// note: if used this no need to push the samples
func FindSizes(cap int, fileSize, chunk int) (size int) {
	if chunk < 1 {
		chunk = 1
	}
	size = max(min(fileSize/chunk, cap), 1)
	return
}

// ComputeAEParams returns the params required for the AEC
// fallback is the cap size set by the udp packet such as 64KB or any if desired not try to push over 64KB
func ComputeAEParams(fallback int, windowDivisor float64, maxMultiplier float64) AEParams {
	//if windowDivisor <= 0 {
	//	windowDivisor = 2
	//}
	//if maxMultiplier <= 0 {
	//	maxMultiplier = 8
	//}

	windowDivisor = max(2, windowDivisor)
	maxMultiplier = max(1, maxMultiplier)

	target := max(fallback, 1)

	w := max(int(float64(target)/windowDivisor), 1)

	maxBoundary := int(float64(target) * maxMultiplier)

	return AEParams{
		TargetSize:  target,
		Window:      w,
		MaxBoundary: maxBoundary,
	}
}

// AEC or Assyemetric Extremum CDC returns the close range of the packet data
// w window size
// lookAhead will be used in the addition make sure to use
// the value which can result the len(chunks) into < udp-cap-packet-level
// ref: https://rickwinfrey.com/assets/papers/cdc/ae-2015-zhang.pdf
func AEC(data []byte, lookAhead uint64, w int) int {

	L := min(len(data), w)
	if L <= 1 {
		return L
	}
	i := 1
	maxValue := data[0]
	maxPos := 0
	for i < L {
		// not doing to use the min|max sutff
		// respecting the formula 🫶
		if data[i] <= maxValue {
			if i == (maxPos + int(lookAhead)) {
				return i
			}
		} else {
			maxValue = data[i]
			maxPos = i
		}
		i += 1
	}
	return L
}

// CreateChunk returns the files into []chnuks
// it uses the Assyemetric Extremum CDC method to chunk the file
// divs: split the chunk into 2D array
// faily honest cuts & w are the most shitest stuff to deal with it
// figure yourself out the fuck going on here
// we believe if it works dont change it 😅
func CreateChunk(data []byte, chunkSize, cuts int, w float64, divs int) IChunk {
	t := time.Now()
	totalLength := len(data)
	target := FindSizes(chunkSize, totalLength, cuts)
	params := ComputeAEParams(target, w, float64(cuts))
	log.Println("computing done: ", time.Since(t))

	estimatedChunks := totalLength / params.TargetSize // by far best invent

	log.Printf("[file-size %s]", ReadSizeString(uint64(totalLength)))
	log.Printf("[params max-boundary %s | target-size %s | window %s]", ReadSizeString(uint64(params.MaxBoundary)), ReadSizeString(uint64(params.TargetSize)), ReadSizeString(uint64(params.Window)))
	log.Printf("[estimatedChunk %s]", ReadSizeString(uint64(estimatedChunks)))

	chunks := make([]IFile, 0, estimatedChunks)
	cursor := 0
	tsize := 0
	var minSize, maxSize = math.MaxInt, 0
	for cursor < totalLength {
		remainingData := data[cursor:totalLength]
		i := max(1, AEC(remainingData, uint64(params.Window), params.MaxBoundary))
		//if i <= 0 {
		//	i = 1
		//}
		chunk := data[cursor : cursor+i] // this is what ace is badass 💀

		// non-sense-stuff
		n := len(chunk)

		minSize = min(n, minSize)
		maxSize = max(n, maxSize)

		//if n < minSize {
		//	minSize = n
		//}
		//if n > maxSize {
		//	maxSize = n
		//}
		tsize += n
		// end

		chunks = append(chunks, IFile{Chunk: chunk})
		cursor += i
	}

	avg := tsize / len(chunks)
	spl := SplitIntoDivisions(chunks, divs)
	log.Printf("[chunk sizes: min=%s avg=%s max=%s]",
		ReadSizeString(uint64(minSize)), ReadSizeString(uint64(avg)), ReadSizeString(uint64(maxSize)))
	if tsize != totalLength {
		log.Printf("[ chunk sizes sum to %d, expected %d]", tsize, totalLength)
	}
	log.Printf("[len %s]", ReadSizeString(uint64(len(chunks))))
	log.Printf("[pieces-count %d]", int(PieceCount(tsize, params.MaxBoundary)))
	return IChunk{
		Files: chunks,
		Split: spl,
	}
}

type IWriteChunk struct {
	MarkExt string
	Err     error
}

// WriteChunkFile writes down the chunks on the disk
func WriteChunkFile(filename string, index int, chunk []byte, src string) *IWriteChunk {
	mext := ".part_"
	w := &IWriteChunk{
		MarkExt: mext,
		Err:     nil,
	}
	absSrc, err := filepath.Abs(src)
	if err != nil {
		w.Err = fmt.Errorf("failed to resolve src: %w", err)
		return w
	}

	if err := os.MkdirAll(absSrc, 0755); err != nil {
		w.Err = err
		return w
	}
	base := filepath.Base(filename)
	ext := fmt.Sprintf("%s%s%d", base, mext, index) // join with this extension
	chunkPath := filepath.Join(absSrc, ext)
	err = os.WriteFile(chunkPath, chunk, 0644)
	if err != nil {
		w.Err = err
		return w
	}
	return w
}

// JoinChunks joins all chunk files of filename from src into store
// if no abs the fuck happening nothing
// or if you are better than me change it with better reason
func JoinChunks(saveAS string, totalChunks int, src, store string) error {
	t, base := time.Now(), filepath.Base(saveAS)

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("[failed to resolve binDir: %w]", err)
	}
	absStore, err := filepath.Abs(store)
	if err != nil {
		return fmt.Errorf("[failed to resolve resultsDir: %w]", err)
	}

	outPath := filepath.Join(absStore, base)

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	outF, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	for i := range totalChunks {
		ext := fmt.Sprintf("%s.part_%d", base, i) // extension
		chunkPath := filepath.Join(absSrc, ext)
		fs, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to read chunk %d at %s: %w", i, chunkPath, err)
		}

		bytesWritten, err := io.Copy(outF, fs)
		fs.Close()
		if err != nil {
			return fmt.Errorf("[copy error %s]", err)
		}
		log.Printf("[JoinChunks] processed chunk %d: %d bytes from %s", i, bytesWritten, chunkPath)
	}

	log.Printf("[JoinChunks] joined %d chunks for %s to %s in %s",
		totalChunks, base, outPath, time.Since(t))
	return nil
}

func SaveChunkToDisk(filename, sDir, bDir string, totalChunks int) error {
	t, base := time.Now(), filepath.Base(filename)
	outPath := filepath.Join(sDir, base)

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	outF, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	for i := range totalChunks {
		chunkPath := filepath.Join(bDir, fmt.Sprintf("%s.part_%d", base, i))
		fs, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to read chunk %d at %s: %w", i, chunkPath, err)
		}
		// joining
		_, err = io.Copy(outF, fs)
		fs.Close()
		if err != nil {
			return fmt.Errorf("[copy error %s]", err)
		}
		// end
	}
	log.Printf("[JoinChunks] joined %d chunks for %s to %s in %s",
		totalChunks, base, outPath, time.Since(t))
	return nil
}
