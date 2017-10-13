package chain

// BlockStore is the interface needed for storing data
type BlockStore interface {
	Get([32]byte) *Block
	Add(Block)
	Length() uint64
	Keys() [][32]byte
	Valid(func([32]byte) bool) bool
}

// MemoryStore is a basic implementation for the physical saving of concrete blocks.
// This POC saves them only in memory
type MemoryStore struct {
	raw []*Block
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
func (b *MemoryStore) Add(block Block) {
	b.raw = append(b.raw, &block)
}

// Length returns the length of the whole chain
func (b *MemoryStore) Length() uint64 {
	return uint64(len(b.raw))
}

// Keys returns a list of hashes of all existing blocks
func (b *MemoryStore) Keys() [][32]byte {
	hkeys := [][32]byte{}
	for _, v := range b.raw {
		hkeys = append(hkeys, v.Hash())
	}
	return hkeys
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
