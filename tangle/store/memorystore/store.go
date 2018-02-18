package memorystore

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"
)

// MemoryStore is a in-memory tangle store
type MemoryStore struct {
	tips map[hash.Hash]bool
	data map[hash.Hash]*site.Site
}

// Init initializes the maps
func (m *MemoryStore) Init(store.Options) error {
	m.tips = make(map[hash.Hash]bool)
	m.data = make(map[hash.Hash]*site.Site)
	return nil
}

// Add adds the record to the data section
func (m *MemoryStore) Add(s *site.Site) error {
	m.data[s.Hash()] = s
	return nil
}

// Get returns the data
func (m *MemoryStore) Get(h hash.Hash) *site.Site {
	return m.data[h]
}

// SetTips applies the delta
func (m *MemoryStore) SetTips(add hash.Hash, del []*site.Site) {
	for _, d := range del {
		delete(m.tips, d.Hash())
	}
	m.tips[add] = true
}

// GetTips returns the tips
func (m *MemoryStore) GetTips() []hash.Hash {
	tips := []hash.Hash{}
	for k := range m.tips {
		tips = append(tips, k)
	}
	return tips
}

// Close does nothing
func (m *MemoryStore) Close() {}

// Size returns the len of the data
func (m *MemoryStore) Size() int {
	return len(m.data)
}

// Hashes returns all stored hashes
func (m *MemoryStore) Hashes() []hash.Hash {
	hs := []hash.Hash{}
	for k := range m.data {
		hs = append(hs, k)
	}
	return hs
}
