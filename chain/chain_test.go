package chain

import (
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

type errorable interface {
	Errorf(string, ...interface{})
	Log(...interface{})
}

func disablelog() {
	log.SetLevel(log.FatalLevel)
}

func emptyChain(s BlockStore) *Chain {
	disablelog()
	c, err := New(s, func([32]byte) bool { return true })
	if err != nil {
		panic(err)
	}
	return c
}

func dummyChain(t errorable, s BlockStore) *Chain {
	c := emptyChain(s)
	for i := 0; i <= 100; i++ {
		_, err := c.Add(Block{PrevHash: c.LastHash(), Content: "Block" + strconv.Itoa(i), Nonce: 0, Date: time.Now()})
		if err != nil {
			t.Errorf("Error adding block to dummy-chain: %s", err)
		}
	}
	return c
}

func TestInitialization(t *testing.T) {
	c := emptyChain(&MemoryStore{})
	if c.Length() != 1 {
		t.Errorf("Invalid Chain Length! Expected %d, got %d", 1, c.Length())
	}
	b := c.Get(c.LastHash())
	if b.Content != "GENESIS" {
		t.Errorf("Invalid genesisBlock content. Got %s", b.Content)
	}
}

func TestAddGet(t *testing.T) {
	c := emptyChain(&MemoryStore{})

	// Adding Block
	lh := c.LastHash()
	subTime := time.Now()
	bh, err := c.Add(Block{PrevHash: lh, Content: "test", Nonce: 0, Date: subTime})
	if err != nil {
		t.Errorf("Error Adding block: %s", err)
	}
	if c.LastHash() == lh {
		t.Errorf("LastHash not updated.")
	}
	if c.LastHash() != bh {
		t.Errorf("LastHash not correct. Expected %x, got %x", bh, c.LastHash())
	}

	// Retrieving block...
	b := c.Get(bh)
	if b == nil {
		t.Error("Unable to get block")
	}
	if !b.Date.Equal(subTime) {
		t.Errorf("Wrong submission time. Expected %s, got %s", subTime.String(), b.Date.String())
	}
	if b.Content != "test" {
		t.Errorf("Wrong content. Expected test, got %s", b.Content)
	}
}

func TestLinking(t *testing.T) {
	c := dummyChain(t, &MemoryStore{})
	b := c.Get(c.LastHash())
	for b.PrevHash != [32]byte{} {
		ob := b
		b = c.Get(b.PrevHash)
		if b == nil {
			t.Errorf("Broken Linkage! %x has no predecessor", ob.Hash())
		}
	}
	if b.Content != "GENESIS" {
		t.Errorf("Invalid Genesis Block. Content was %s", b.Content)
	}
}

func TestRandomAccess(t *testing.T) {
	c := dummyChain(t, &MemoryStore{})
	b := c.Get(c.LastHash())
	for i := 0; i < 50; i++ {
		b = c.Get(b.PrevHash)
	}
	if b.Content != "Block50" {
		t.Errorf("Invalid Block Content! Expected Block50, got %s", b.Content)
	}
}
