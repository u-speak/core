package chain

import (
	"errors"
)

// MemoryStore is a basic implementation for the physical saving of concrete blocks.
// This POC saves them only in memory
type MemoryStore struct {
	raw         []*Block
	initialized bool
}

// Init initializes the raw storage
func (b *MemoryStore) Init() ([32]byte, error) {
	if len(b.raw) == 0 {
		bl := genesisBlock()
		b.raw = append(b.raw, &bl)
	}
	preds := make(map[[32]byte]bool)
	for _, b := range b.raw {
		preds[b.PrevHash] = true
	}
	for _, b := range b.raw {
		if preds[b.Hash()] == false {
			return b.Hash(), nil
		}
	}
	b.initialized = true
	return [32]byte{}, errors.New("Could not find lasthash")
}

// Initialized returns whether or not this store has been initialized
func (b *MemoryStore) Initialized() bool {
	return b.initialized
}

// Get retrieves a block by its hash
func (b *MemoryStore) Get(hash [32]byte) *Block {
	for i := range b.raw {
		if b.raw[i].Hash() == hash {
			return b.raw[i]
		}
	}
	return nil
}

// Add adds a block to the raw storage
func (b *MemoryStore) Add(block Block) error {
	b.raw = append(b.raw, &block)
	return nil
}

// Length returns the length of the whole chain
func (b *MemoryStore) Length() uint64 {
	return uint64(len(b.raw))
}

func (b *MemoryStore) bloomFilter() map[[32]byte]bool {
	f := make(map[[32]byte]bool)
	for _, v := range b.raw {
		f[v.Hash()] = true
	}
	return f
}

// Valid checks if all blocks are connected and have the required difficulty
func (b *MemoryStore) Valid(v func([32]byte) bool) bool {
	f := b.bloomFilter()
	for _, b := range b.raw {
		if !v(b.Hash()) {
			return false
		} else if !f[b.PrevHash] && b.Content != "GENESIS" {
			return false
		}
	}
	return true
}

// Close closes the underlying connections
func (b *MemoryStore) Close() {
}

// Reinitialize resets the chain
func (b *MemoryStore) Reinitialize() ([32]byte, error) {
	b.raw = []*Block{}
	return b.Init()
}
