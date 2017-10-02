package chain

// BlockStore is the interface needed for storing data
type BlockStore interface {
	Get([32]byte) *Block
	Add(*Block)
	Length() uint64
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
func (b *MemoryStore) Add(block *Block) {
	b.raw = append(b.raw, block)
}

// Length returns the length of the whole chain
func (b *MemoryStore) Length() uint64 {
	return uint64(len(b.raw))
}
