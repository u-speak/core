package chain

import (
	"github.com/coreos/bbolt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

var boltPath = path.Join(os.TempDir(), "uspeak-test.db")

func emptyBoltStore() *BoltStore {
	disablelog()
	_ = os.Remove(boltPath)
	s := &BoltStore{Path: boltPath}
	return s
}

func TestBoltStoreInitialization(t *testing.T) {
	disablelog()
	_ = os.Remove(boltPath)
	s := &BoltStore{Path: boltPath}
	lh, err := s.Init()
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}
	if s.Length() != 1 {
		t.Errorf("Invalid Count of stored Blocks! Expected %d, got %d", 1, s.Length())
	}
	g := genesisBlock()
	if lh != g.Hash() {
		t.Errorf("Last hash calculation incorrect")
	}
	b := s.Get(lh)
	if b.Content != "GENESIS" {
		t.Errorf("Invalid genesisBlock content. Got %s", b.Content)
	}
}

func TestBoltStoreAddGet(t *testing.T) {
	s := emptyBoltStore()
	_, _ = s.Init()
	subTime := time.Now()
	block := Block{Content: "test", Nonce: 0, Date: subTime}
	err := s.Add(block)
	if err != nil {
		t.Errorf("Error Adding block: %s", err)
	}

	// Retrieving block...
	b := s.Get(block.Hash())
	if b == nil {
		t.Error("Unable to get block")
		return
	}
	if !b.Date.Equal(subTime) {
		t.Errorf("Wrong submission time. Expected %s, got %s", subTime.String(), b.Date.String())
	}
	if b.Content != "test" {
		t.Errorf("Wrong content. Expected test, got %s", b.Content)
	}
}

func TestBoltStoreValidation(t *testing.T) {
	disablelog()
	c := dummyChain(t, emptyBoltStore())
	if !c.Valid() {
		t.Error("Validation of valid chain failed")
	}
	h, err := c.Add(Block{PrevHash: c.LastHash(), Content: "foo"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !c.Valid() {
		t.Error("Validation of valid chain failed")
	}

	s := c.blocks.(*BoltStore)
	err = s.db.Update(func(tx *bolt.Tx) error {
		blocks := tx.Bucket([]byte("blocks"))
		mod := &Block{PrevHash: c.LastHash(), Content: "modified"}
		enc, err := mod.encode()
		if err != nil {
			return err
		}
		return blocks.Put(h[:], enc)
	})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if c.Valid() {
		t.Error("Modified Chain reported as valid")
	}
}

func TestBoltStoreRestore(t *testing.T) {
	p := path.Join(os.TempDir(), "uspeak-test-restore.db")
	_ = os.Remove(boltPath)
	s := &BoltStore{Path: p}
	c1, err := New(s, func(Hash) bool { return true })
	if err != nil {
		t.Errorf("Chain creation failed: %s", err)
	}
	b1 := Block{PrevHash: c1.LastHash(), Content: "Block", Nonce: 0, Date: time.Now()}
	h := b1.Hash()
	_, err = c1.Add(b1)
	if err != nil {
		t.Error(err)
	}
	s.Close()
	c2, err := New(s, func(Hash) bool { return true })
	if c2.Length() != c1.Length() {
		t.Error("Chains are not the same length")
	}
	b2 := c2.Get(h)
	if !(b2 != nil && b2.Content == "Block") {
		t.Error("Chains did not contain the same block")
	}
}

func BenchmarkBoltStoreAdd(b *testing.B) {
	s := emptyBoltStore()
	_, _ = s.Init()
	for i := 0; i < b.N; i++ {
		err := s.Add(Block{Content: "Block" + strconv.Itoa(i), Nonce: 0, Date: time.Now()})
		if err != nil {
			b.Errorf("Error adding Block: %s", err)
		}
	}
}

func BenchmarkBoltStoreGet(b *testing.B) {
	c := dummyChain(b, emptyBoltStore())
	ks := c.blocks.(*BoltStore).keys()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := rand.Intn(len(ks))
		c.Get(ks[k])
	}
}

func BenchmarkBoltStoreValidation(b *testing.B) {
	c := dummyChain(b, emptyBoltStore())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !c.Valid() {
			b.Error("Chain validation failed")
		}
	}
}
