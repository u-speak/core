package util

import (
	"github.com/martinlindhe/bubblebabble"
)

// EncodeBubbleBabble is a wrapper function to encode hashes into a human readable format
func EncodeBubbleBabble(h [32]byte) string {
	dst := make([]byte, bubblebabble.EncodedLen(32))
	bubblebabble.Encode(dst, h[:])
	return string(dst)
}

// DecodeBubbleBabble is a wrapper function to decode hashes from a human readable format
func DecodeBubbleBabble(s string) ([32]byte, error) {
	dst := [32]byte{}
	_, err := bubblebabble.Decode(dst[:], []byte(s))
	return dst, err
}
