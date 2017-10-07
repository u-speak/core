package chain

import (
	"crypto/sha256"
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

func (b *Block) Hash() [32]byte {
	return sha256.Sum256(append([]byte(b.Signature+b.PubKey+string(b.Nonce)+b.Content+string(b.Date.Unix())), b.PrevHash[:]...))
}

func genesisBlock() *Block {
	return &Block{
		Nonce:     0,
		PrevHash:  [32]byte{},
		Content:   "GENESIS",
		Signature: "",
		PubKey:    "",
		Type:      "GENESIS",
		Date:      time.Unix(0, 0),
	}
}
