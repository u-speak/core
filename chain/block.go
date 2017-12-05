package chain

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/vmihailenco/msgpack"
	"strconv"
	"time"
)

// Hash is a wrapper to easily compare digests
type Hash [32]byte

// A Block is a concrete data entity. It stores the content as well as metadata
// TODO: Maybe separate content to minimize memory usage
type Block struct {
	Nonce     uint32
	PrevHash  Hash
	Content   string
	Signature string
	Type      string
	PubKey    string
	Date      time.Time
}

// Hash returns the sha256 hash of the block
func (b Block) Hash() Hash {
	// Interject the bstr with literals to prevent attacks on the block structure
	bstr := "C" + b.Content + "T" + b.Type + "S" + b.Signature + "P" + b.PubKey + "D" + strconv.FormatUint(uint64(b.Date.Unix()), 10) + "N" + strconv.FormatUint(uint64(b.Nonce), 10) + "PREV" + base64.URLEncoding.EncodeToString(b.PrevHash[:])
	return sha256.Sum256([]byte(bstr))
}

func (b *Block) encode() ([]byte, error) {
	return msgpack.Marshal(b)
}

// DecodeBlock decodes a byte array back to a block
func DecodeBlock(data []byte) (*Block, error) {
	b := &Block{}
	err := msgpack.Unmarshal(data, b)
	return b, err
}

func genesisBlock() Block {
	return Block{
		Nonce:     0,
		PrevHash:  [32]byte{},
		Content:   "GENESIS",
		Signature: "",
		PubKey:    "",
		Type:      "GENESIS",
		Date:      time.Unix(0, 0),
	}
}

// EqSlice is a utility function to compare the hash to a slice instead of a fixed size array
func (h Hash) EqSlice(s []byte) bool {
	var c Hash
	copy(c[:], s)
	return h == c
}

// Empty returns whether the hash is empty
func (h Hash) Empty() bool {
	return h == [32]byte{}
}

// FromSlice returns a hash from a slice
func FromSlice(s []byte) Hash {
	var h Hash
	copy(h[:], s)
	return h
}

// HasPrefix returns true when the slice is the prefix of the hash
func (h Hash) HasPrefix(s []byte) bool {
	if len(s) >= len(h) {
		return h.EqSlice(s)
	}
	for i, b := range s {
		if b != h[i] {
			return false
		}
	}
	return true
}

// Bytes returns the hash as a slice
func (h Hash) Bytes() []byte {
	return h[:]
}
