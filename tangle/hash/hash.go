package hash

import (
	"encoding/base64"
	"github.com/deckarep/golang-set"
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

// Diff returns the difference between the local and remote hashes
func Diff(l, r []Hash) ([]Hash, []Hash) {
	loc := mapset.NewSet()
	for _, h := range l {
		loc.Add(h)
	}
	rem := mapset.NewSet()
	for _, h := range r {
		rem.Add(h)
	}
	delm := loc.Difference(rem)
	addm := rem.Difference(loc)
	a, d := []Hash{}, []Hash{}

	for _, h := range delm.ToSlice() {
		d = append(d, h.(Hash))
	}
	for _, h := range addm.ToSlice() {
		a = append(a, h.(Hash))
	}
	return a, d
}

// FromSlice turns a byte slice into a hash
func FromSlice(s []byte) Hash {
	h := Hash{}
	copy(h[:], s)
	return h
}
