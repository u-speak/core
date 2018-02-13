package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/u-speak/core/tangle/hash"
	"golang.org/x/crypto/blake2s"
)

var dummyContent = blake2s.Sum256([]byte{1, 3, 3, 7})

var dummySite = Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&Site{Content: dummyContent, Nonce: 0}}}
var complexSite = Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&dummySite, &Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&dummySite}}}}

func TestHash(t *testing.T) {
	// Testing single site
	s := &Site{Content: dummyContent, Nonce: 0}
	h := s.Hash()
	assert.Equal(t, hash.Hash{0xcd, 0x5b, 0xe2, 0x88, 0x5d, 0xb7, 0xa3, 0xd6, 0xbe, 0x45, 0xd7, 0x36, 0x4d, 0xa3, 0x28, 0x86, 0x5d, 0xf, 0x1a, 0xbf, 0x3d, 0x9f, 0xfd, 0x48, 0x38, 0xf5, 0x7, 0x7e, 0x31, 0xb1, 0xa3, 0x82}, h)

	// Testing linked sites
	assert.Equal(t, hash.Hash{0xf6, 0x9e, 0x1a, 0x6f, 0x3c, 0xaa, 0xe, 0xe0, 0x29, 0x2e, 0x4f, 0xb2, 0xa5, 0x32, 0x38, 0x5b, 0xe5, 0x99, 0xb4, 0x45, 0x9e, 0xa4, 0x3b, 0xc, 0x90, 0x11, 0x87, 0xde, 0x89, 0x40, 0x92, 0xde}, dummySite.Hash())
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
