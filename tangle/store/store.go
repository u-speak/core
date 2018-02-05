package store

import (
	"github.com/u-speak/core/tangle/hash"
)

// Serializable structs can be turned into slices and also be recovered from them
type Serializable interface {
	Serialize() []byte
	Deserialize([]byte) error
	Hash() hash.Hash
}

// Store is a persistant datastore
type Store interface {
	Add(Serializable) error
	Get(hash.Hash) []byte
	Init(Options) error
	Close()
}

// Options for the store, used at initialization
type Options struct {
	Path string
}
