package chain

import (
	"crypto/sha256"
	"strconv"
	"time"
)

type Block struct {
	Nonce     uint
	PrevHash  [32]byte
	Content   string
	Signature string
	Type      string
	PubKey    string
	Date      time.Time
}

func (b Block) Hash() [32]byte {
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
