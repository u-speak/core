package chain

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type errorable interface {
	Errorf(string, ...interface{})
}

func emptyChain() *Chain {
	s := &MemoryStore{}
	return New(s, func([32]byte) bool { return true })
}

func dummyChain(t errorable) *Chain {
	c := emptyChain()
	for i := 0; i <= 100; i++ {
		_, err := c.Add(Block{PrevHash: c.LastHash(), Content: "Block" + strconv.Itoa(i), Nonce: 0, Date: time.Now()})
		if err != nil {
			t.Errorf("Error adding block to dummy-chain: %s", err)
		}
	}
	return c
}

func TestInitialization(t *testing.T) {
	c := emptyChain()
	if c.Length() != 1 {
		t.Errorf("Invalid Chain Length! Expected %d, got %d", 1, c.Length())
	}
	b := c.Get(c.LastHash())
	if b.Content != "GENESIS" {
		t.Errorf("Invalid genesisBlock content. Got %s", b.Content)
	}
}

func TestAddGet(t *testing.T) {
	c := emptyChain()

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
	if b.Date != subTime {
		t.Errorf("Wrong submission time. Expected %s, got %s", subTime.String(), b.Date.String())
	}
	if b.Content != "test" {
		t.Errorf("Wrong content. Expected test, got %s", b.Content)
	}
}

func TestLinking(t *testing.T) {
	c := dummyChain(t)
	b := c.Get(c.LastHash())
	for b.PrevHash != [32]byte{} {
		ob := b
		b = c.Get(b.PrevHash)
		if b == nil {
			t.Errorf("Broken Linkage! %s has no predecessor", ob)
		}
	}
	if b.Content != "GENESIS" {
		t.Errorf("Invalid Genesis Block. Content was %s", b.Content)
	}
}

func TestRandomAccess(t *testing.T) {
	c := dummyChain(t)
	b := c.Get(c.LastHash())
	for i := 0; i < 50; i++ {
		b = c.Get(b.PrevHash)
	}
	if b.Content != "Block50" {
		t.Errorf("Invalid Block Content! Expected Block50, got %s", b.Content)
	}
}

func TestValidation(t *testing.T) {
	c := dummyChain(t)
	if !c.Valid() {
		t.Error("Validation of valid chain failed")
	}
	c.blocks.(*MemoryStore).raw[4].Content = "MODIFIED"
	if c.Valid() {
		t.Error("Modified Chain reported as valid")
	}
}

func BenchmarkAdd(b *testing.B) {
	c := emptyChain()
	for i := 0; i < b.N; i++ {
		_, err := c.Add(Block{PrevHash: c.LastHash(), Content: "Block" + strconv.Itoa(i), Nonce: 0, Date: time.Now()})
		if err != nil {
			b.Errorf("Error adding Block: %s", err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	c := dummyChain(b)
	b.ResetTimer()
	keys := c.blocks.Keys()
	for i := 0; i < b.N; i++ {
		k := rand.Intn(len(keys))
		c.Get(keys[k])
	}
}

func BenchmarkValidation(b *testing.B) {
	c := dummyChain(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !c.Valid() {
			b.Error("Chain validation failed")
		}
	}
}
