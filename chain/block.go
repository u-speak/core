package chain

import (
	"crypto/sha256"
)

type Block struct {
	Nonce     uint
	PrevHash  [32]byte
	Content   string
	Signature string
	Type      string
	PubKey    string
}

func (b *Block) Hash() [32]byte {
	return sha256.Sum256(append([]byte(b.Signature+b.PubKey+string(b.Nonce)+b.Content), b.PrevHash[:]...))
}
