package chain

import (
	"crypto/sha256"
	"strconv"
	"time"
)

// A Block is a concrete data entity. It stores the content as well as metadata
// TODO: Maybe separate content to minimize memory usage
type Block struct {
	Nonce     uint
	PrevHash  [32]byte
	Content   string
	Signature string
	Type      string
	PubKey    string
	Date      time.Time
}

// Hash returns the sha256 hash of the block
func (b Block) Hash() [32]byte {
	// Interject the bstr with literals to prevent attacks on the block structure
	bstr := "C" + b.Content + "T" + b.Type + "S" + b.Signature + "P" + b.PubKey + "D" + strconv.FormatUint(uint64(b.Date.Unix()), 10) + "N" + string(b.Nonce) + "PREV"
	return sha256.Sum256(append([]byte(bstr), b.PrevHash[:]...))
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
