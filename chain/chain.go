package chain

import (
	"errors"
)

// ValidationFunc is the requirement for mining
type ValidationFunc func([32]byte) bool

// Chain is a Blockchain Implementation
type Chain struct {
	blocks   BlockStore
	lastHash [32]byte
	validate ValidationFunc
}

// New initializes a new Chain
func New(b BlockStore, validate ValidationFunc) *Chain {
	g := genesisBlock()
	b.Add(g)
	c := &Chain{blocks: b, validate: validate}
	c.lastHash = g.Hash()
	return c
}

func (c *Chain) Add(b Block) ([32]byte, error) {
	hash := b.Hash()
	if !c.validate(hash) {
		return [32]byte{}, errors.New("Block did not pass the validation function")
	}
	if b.PrevHash != c.lastHash {
		return [32]byte{}, errors.New("Blocks PrevHash was not the lasthash")
	}
	c.blocks.Add(b)
	c.lastHash = hash
	return hash, nil
}

// DumpChain dumps the whole ordered chain in an array
func (c *Chain) DumpChain() ([]*Block, error) {
	if !c.IsValid() {
		return []*Block{}, errors.New("Chain is not Valid! Cannot dump")
	}
	h := c.lastHash
	bl := []*Block{}
	for h != [32]byte{} {
		b := c.Get(h)
		bl = append(bl, b)
		h = b.PrevHash
	}
	return bl, nil
}

// Get retrieves a block
func (c *Chain) Get(hash [32]byte) *Block {
	return c.blocks.Get(hash)
}

// IsValid checks the chain for integrity and validation compliance
func (c *Chain) IsValid() bool {
	if c.lastHash == [32]byte{} {
		return true
	}
	b := c.blocks.Get(c.lastHash)
	for b != nil {
		if b.PrevHash == [32]byte{} {
			return true
		}
		if !c.validate(b.Hash()) {
			return false
		}
		b = c.blocks.Get(b.PrevHash)
	}
	return false
}

// LastHash returns the hash of the last block in the chain
func (c *Chain) LastHash() [32]byte {
	return c.lastHash
}

// Length returns the length of the whole chain
func (c *Chain) Length() uint64 {
	return c.blocks.Length()
}
