package boltstore

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/store"

	"github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
)

var (
	bucketName = []byte("data")
)

// BoltStore stores its persistence data in a boltdb (github.com/coreos/bbolt)
type BoltStore struct {
	db *bolt.DB
}

// Add stores the data in the database
func (b *BoltStore) Add(d store.Serializable) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketName)
		return bkt.Put(d.Hash().Slice(), d.Serialize())
	})
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves data from the database
func (b *BoltStore) Get(h hash.Hash) []byte {
	var d []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketName)
		d = bkt.Get(h.Slice())
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return d
}

// Init the store
func (b *BoltStore) Init(o store.Options) error {
	db, err := bolt.Open(o.Path, 0444, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	b.db = db
	return nil
}

// Close releases the lock on the db
func (b *BoltStore) Close() {
	err := b.db.Close()
	if err != nil {
		log.Error(err)
	}
}
