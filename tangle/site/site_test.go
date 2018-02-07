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
	assert.Equal(t, hash.Hash{0xda, 0x49, 0xae, 0x6b, 0x16, 0x36, 0x91, 0x36, 0x51, 0xf1, 0x5f, 0x24, 0x96, 0x75, 0xc6, 0xf3, 0x9f, 0x51, 0x6, 0x8c, 0x91, 0x5d, 0xba, 0xed, 0x98, 0xe8, 0x4f, 0x44, 0xda, 0x49, 0x8e, 0x98}, h)

	// Testing linked sites
	assert.Equal(t, hash.Hash{0x5b, 0x7b, 0x74, 0x87, 0x6e, 0x63, 0xd2, 0x16, 0x29, 0x2e, 0xae, 0x33, 0x3f, 0x49, 0xf7, 0xa3, 0x3b, 0x53, 0xcd, 0x64, 0x14, 0x85, 0x9f, 0x6a, 0x26, 0xb2, 0x86, 0x13, 0x83, 0xff, 0xad, 0x78}, dummySite.Hash())
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
