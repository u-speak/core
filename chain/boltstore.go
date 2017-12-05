package chain

import (
	"encoding/base64"
	"os"

	"github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
)

// BoltStore is a Blockstore implementation using the BoltDB key value store
type BoltStore struct {
	Path        string
	db          *bolt.DB
	initialized bool
}

// Init initializes the BoltStore
func (b *BoltStore) Init() (Hash, error) {
	if b.initialized {
		return Hash{}, ErrStoreInitialized
	}
	db, err := bolt.Open(b.Path, 0600, nil)
	if err != nil {
		return Hash{}, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("meta"))
		if err != nil {
			return err
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte("blocks"))
		if err != nil {
			return err
		}
		if bucket.Stats().KeyN == 0 {
			log.Infof("Initializing empty chain at path %s", b.Path)
			g := genesisBlock()
			h := g.Hash()
			enc, err := g.encode()
			if err != nil {
				return err
			}
			err = bucket.Put(h[:], enc)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Hash{}, err
	}
	var lh Hash
	err = db.View(func(tx *bolt.Tx) error {
		meta := tx.Bucket([]byte("meta"))
		hs := meta.Get([]byte("lasthash"))
		if hs != nil {
			log.WithField("db", b.Path).Info("Using stored LastHash")
			copy(lh[:], hs)
			return nil
		}
		log.WithField("db", b.Path).Info("No metadata saved, calculating lasthash")
		blocks := tx.Bucket([]byte("blocks"))
		bloom := map[Hash]bool{}
		err := blocks.ForEach(func(k, v []byte) error {
			bl, err := DecodeBlock(v)
			if err != nil {
				return err
			}
			bloom[bl.PrevHash] = true
			return nil
		})
		if err != nil {
			return err
		}
		var kh Hash
		err = blocks.ForEach(func(k, v []byte) error {
			copy(kh[:], k)
			if bloom[kh] == false {
				lh = kh
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	err = db.Update(func(tx *bolt.Tx) error {
		meta := tx.Bucket([]byte("meta"))
		return meta.Put([]byte("lasthash"), lh[:])
	})
	if err != nil {
		return Hash{}, err
	}
	log.WithFields(log.Fields{
		"db":       b.Path,
		"LastHash": base64.URLEncoding.EncodeToString(lh[:]),
	}).Infof("Finished initialization")
	b.db = db
	b.initialized = true
	return lh, nil
}

// Initialized returns whether or not this store has been initialized
func (b *BoltStore) Initialized() bool {
	return b.initialized
}

// Get retrieves a block by its hash
func (b *BoltStore) Get(hash Hash) *Block {
	var bl *Block
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		bs := bucket.Get(hash[:])
		bb, err := DecodeBlock(bs)
		bl = bb
		return err
	})
	if err != nil {
		return nil
	}
	return bl
}

// Add adds a block to the raw storage
func (b *BoltStore) Add(block Block) error {
	h := block.Hash()
	enc, err := block.encode()
	if err != nil {
		return err
	}
	err = b.db.Update(func(tx *bolt.Tx) error {
		blocks := tx.Bucket([]byte("blocks"))
		err := blocks.Put(h[:], enc)
		if err != nil {
			return err
		}

		meta := tx.Bucket([]byte("meta"))
		err = meta.Put([]byte("lasthash"), h[:])
		return err
	})
	return err
}

func (b *BoltStore) keys() []Hash {
	ret := []Hash{}
	_ = b.db.View(func(tx *bolt.Tx) error {
		blocks := tx.Bucket([]byte("blocks"))
		var kh Hash
		_ = blocks.ForEach(func(k, _ []byte) error {
			copy(kh[:], k)
			ret = append(ret, kh)
			return nil
		})
		return nil
	})
	return ret
}

// Length returns the length of the whole chain
func (b *BoltStore) Length() uint64 {
	count := 0
	_ = b.db.View(func(tx *bolt.Tx) error {
		blocks := tx.Bucket([]byte("blocks"))
		count = blocks.Stats().KeyN
		return nil
	})
	return uint64(count)
}

// Valid checks if all blocks are connected and have the required difficulty
func (b *BoltStore) Valid(val func(Hash) bool) bool {
	if b.Length() == 0 {
		return false
	}
	valid := true
	f := make(map[Hash]bool)
	err := b.db.View(func(tx *bolt.Tx) error {
		blocks := tx.Bucket([]byte("blocks"))
		var idx Hash
		_ = blocks.ForEach(func(k, v []byte) error {
			copy(idx[:], k)
			f[idx] = true
			return nil
		})
		c := blocks.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			b, err := DecodeBlock(v)
			if err != nil {
				return err
			}
			h := b.Hash()
			copy(idx[:], k)
			if idx != h {
				valid = false
				break
			} else if !f[b.PrevHash] && b.Content != "GENESIS" {
				valid = false
				break
			} else if !val(h) {
				valid = false
				break
			}
		}
		return nil
	})
	if err != nil {
		return false
	}
	return valid
}

// Close closes the underlying connections
func (b *BoltStore) Close() {
	_ = b.db.Close()
	b.initialized = false
}

// Reinitialize resets the chain
func (b *BoltStore) Reinitialize() (Hash, error) {
	b.Close()
	_ = os.Remove(b.Path)
	return b.Init()
}

// ApproxHash returns the approximated hash
func (b *BoltStore) ApproxHash(s []byte) Hash {
	var h Hash
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		bucket.ForEach(func(k, v []byte) error {
			if FromSlice(k).HasPrefix(s) {
				h = FromSlice(k)
				return nil
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return [32]byte{}
	}
	return [32]byte{}
}
