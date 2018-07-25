package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/u-speak/core/tangle/hash"
	"golang.org/x/crypto/blake2b"
)

var dummyContent = blake2b.Sum256([]byte{1, 3, 3, 7})

var dummySite = Site{Content: dummyContent, Nonce: 0, Validates: []*Site{{Content: dummyContent, Nonce: 0}}}
var complexSite = Site{Content: dummyContent, Nonce: 0, Validates: []*Site{&dummySite, {Content: dummyContent, Nonce: 0, Validates: []*Site{&dummySite}}}}

func TestHash(t *testing.T) {
	// Testing single site
	s := &Site{Content: dummyContent, Nonce: 0}
	h := s.Hash()
	assert.Equal(t, hash.Hash{0x63, 0x92, 0xa7, 0x4, 0x4, 0x4c, 0xfc, 0x90, 0x42, 0xfc, 0xb9, 0x36, 0xea, 0xde, 0xdb, 0xa4, 0xbd, 0xc8, 0xb5, 0xcf, 0x2d, 0x17, 0x69, 0x3, 0x4e, 0x6f, 0xad, 0x7b, 0xf, 0xfa, 0x0, 0xe1}, h)

	// Testing linked sites
	assert.Equal(t, hash.Hash{0x8c, 0x98, 0xc5, 0x7d, 0xb8, 0x78, 0x76, 0x8c, 0xe8, 0xcf, 0xb, 0x2e, 0xfb, 0xfa, 0x9a, 0x69, 0xf, 0x6d, 0x77, 0xe5, 0x16, 0x9e, 0x29, 0xa6, 0x41, 0x44, 0x6a, 0x27, 0x74, 0x52, 0xae, 0x55}, dummySite.Hash())
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
