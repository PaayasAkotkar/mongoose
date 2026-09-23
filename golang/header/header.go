package header

import (
	"fmt"
	"sort"
	"strings"
)

type IHeader struct {
	// check the transfer NAT network platform
	IPv4 bool
	IPv6 bool
	// end
	cache   map[string]string
	headers string
}

func NewHeader() *IHeader {
	return &IHeader{
		cache: make(map[string]string),
	}
}

func (h *IHeader) SetHeader(key string, value string) *IHeader {
	if strings.ContainsAny(key, ":\n") || strings.ContainsAny(value, ":\n") {
		fmt.Printf("key/value cannot contain ':' or newline: %q=%q", key, value)
		return h
	}
	if h.cache == nil {
		h.cache = make(map[string]string)
	}
	h.cache[key] = value
	return h
}

// Build composes the set-header into string readable way
func (h *IHeader) Build() {
	keys := make([]string, 0, len(h.cache))
	for k := range h.cache {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&sb, "%s:%s\n", k, h.cache[k])
	}
	h.headers = sb.String() // note: sb is a strings.Builder, see Bug 3
}

func (h *IHeader) Read() string {
	return h.headers
}

// Release imp for cleanup after build
// it simply deletes the map
func (h *IHeader) Release() {
	for k := range h.cache {
		delete(h.cache, k)
	}
}
