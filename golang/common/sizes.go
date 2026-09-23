package common

import (
	"github.com/google/uuid"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

// GenerateHash generates a random hash for temporary operations
func GenerateHash() string {
	return uuid.New().String()
}
