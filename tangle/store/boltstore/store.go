package boltstore

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"

	"github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
)

var (
	dataBucketName = []byte("data")
	tipBucketName  = []byte("tips")
)

// BoltStore stores its persistence data in a boltdb (github.com/coreos/bbolt)
type BoltStore struct {
	db *bolt.DB
}

// New returns a fresh initialized store
func New(o store.Options) (*BoltStore, error) {
	s := &BoltStore{}
	return s, s.Init(o)
}

// Add stores the data in the database
func (b *BoltStore) Add(d *site.Site) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(dataBucketName)
		return bkt.Put(d.Hash().Slice(), d.Serialize())
	})
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves data from the database
func (b *BoltStore) Get(h hash.Hash) *site.Site {
	var d []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(dataBucketName)
		d = bkt.Get(h.Slice())
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	if d == nil {
		return nil
	}
	s := site.Site{}
	err = s.Deserialize(d)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &s
}

// Init the store
func (b *BoltStore) Init(o store.Options) error {
	db, err := bolt.Open(o.Path, 0644, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(dataBucketName)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(tipBucketName)
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

// SetTips applies the delata of tips
func (b *BoltStore) SetTips(add hash.Hash, del []*site.Site) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipBucketName)
		for _, d := range del {
			err := bkt.Delete(d.Hash().Slice())
			if err != nil {
				return err
			}
		}
		err := bkt.Put(add.Slice(), []byte{})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
}

// GetTips returns the saved tips
func (b *BoltStore) GetTips() []hash.Hash {
	tips := []hash.Hash{}
	_ = b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(tipBucketName)
		_ = bkt.ForEach(func(k []byte, _ []byte) error {
			tips = append(tips, hash.FromSlice(k))
			return nil
		})
		return nil
	})
	return tips
}

// Size returns the count of elements in the data bucket
func (b *BoltStore) Size() int {
	var n int
	_ = b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(dataBucketName)
		n = bkt.Stats().KeyN
		return nil
	})
	return n
}

// Hashes returns all stored hashes
func (b *BoltStore) Hashes() []hash.Hash {
	hs := []hash.Hash{}
	_ = b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(dataBucketName)
		_ = bkt.ForEach(func(k, _ []byte) error {
			hs = append(hs, hash.FromSlice(k))
			return nil
		})
		return nil
	})
	return hs
}
