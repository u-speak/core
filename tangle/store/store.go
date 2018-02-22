package store

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
)

// Store is a persistant datastore
type Store interface {
	Add(*site.Site) error
	Get(hash.Hash) *site.Site
	Init(Options) error
	SetTips(hash.Hash, []*site.Site)
	GetTips() []hash.Hash
	Hashes() []hash.Hash
	Size() int
	Close()
}

// Empty checks whether this store has been used before
func Empty(s Store) bool {
	return len(s.GetTips()) == 0
}

// Options for the store, used at initialization
type Options struct {
	Path string
}
