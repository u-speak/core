package datastore

import (
	"github.com/coreos/bbolt"
	"github.com/u-speak/core/tangle/hash"
)

var (
	bucketname = []byte("data")
)

// Serializable allows for the storage of any kind of data
type Serializable interface {
	Hash() (hash.Hash, error)
	Serialize() ([]byte, error)
	Deserialize([]byte) error
	Type() string
}

// Store is responsible for storing the actual data on the tangle
type Store struct {
	db *bolt.DB
}

// New returns an initialized Store
func New(path string) (*Store, error) {
	s := &Store{}
	db, err := bolt.Open(path, 0644, nil)
	if err != nil {
		return nil, err
	}
	s.db = db
	err = s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketname)
		return err
	})
	return s, err
}

// Put stores the serialized element in the database
func (s *Store) Put(e Serializable) error {
	h, err := e.Hash()
	if err != nil {
		return err
	}
	d, err := e.Serialize()
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketname).Put(h.Slice(), d)
	})
}

// Get retrieves the serialized object
func (s *Store) Get(dest Serializable, h hash.Hash) error {
	var buff []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		buff = tx.Bucket(bucketname).Get(h.Slice())
		return nil
	})
	if err != nil {
		return err
	}
	return dest.Deserialize(buff)
}
