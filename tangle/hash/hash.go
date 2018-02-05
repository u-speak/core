package hash

import (
	"encoding/base64"
)

const (
	// HashSize of the stored hash
	HashSize = 32
)

// Hash is a wrapper around blake2s
type Hash [HashSize]byte

func (h Hash) String() string {
	return base64.StdEncoding.EncodeToString(h[:])
}

// Weight is the difficulty (or number of leading zeroes) of a site
func (h Hash) Weight() int {
	weight := 0
	for _, b := range h {
		if b == 0 {
			weight++
		} else {
			return weight
		}
	}
	return weight
}

// Slice converts the fixed length hash to a dynamic slice
func (h Hash) Slice() []byte {
	return h[:]
}
