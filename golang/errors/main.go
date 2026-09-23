// Package merror implements the typically string written errors
package merror

import (
	"errors"
)

var (
	NilError     = errors.New("[nil-error]")
	AcceeptError = errors.New("[accept-error %s]")
)
