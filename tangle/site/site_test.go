package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/u-speak/core/tangle/hash"
	"golang.org/x/crypto/blake2b"
)

var dummyContent = blake2b.Sum256([]byte{1, 3, 3, 7})

var dummySite = Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&Site{Content: dummyContent, Nonce: 0}}}
var complexSite = Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&dummySite, &Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&dummySite}}}}

func TestHash(t *testing.T) {
	// Testing single site
	s := &Site{Content: dummyContent, Nonce: 0}
	h := s.Hash()
	assert.Equal(t, hash.Hash{0x9d, 0xfb, 0x50, 0x40, 0xc1, 0x12, 0x7b, 0xbb, 0x7e, 0xca, 0xbe, 0x50, 0x89, 0xe, 0x32, 0x67, 0xe6, 0xf0, 0x51, 0x3b, 0xce, 0xfb, 0x7f, 0x98, 0x99, 0x14, 0x8d, 0x33, 0xa9, 0xe0, 0x9b, 0x20}, h)

	// Testing linked sites
	assert.Equal(t, hash.Hash{0x55, 0x8e, 0x52, 0x9f, 0xe6, 0x35, 0xb4, 0xc8, 0xd9, 0x60, 0xaf, 0xc7, 0xf3, 0x70, 0x7, 0x36, 0x64, 0x3b, 0xb, 0xad, 0xb9, 0x15, 0x9b, 0x42, 0x25, 0xe3, 0x47, 0xe0, 0xc, 0xf7, 0x2d, 0x57}, dummySite.Hash())
}

func BenchmarkSimpleSite(b *testing.B) {
	s := &Site{Content: dummyContent, Nonce: 0}
	for i := 0; i < b.N; i++ {
		s.Hash()
	}
}

func BenchmarkDummySite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dummySite.Hash()
	}
}

func BenchmarkComplexSite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		complexSite.Hash()
	}
}
