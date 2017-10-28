package chain

import (
	"encoding/base64"
	"encoding/gob"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var testPath = path.Join(os.TempDir(), "uspeak-core-test")

type errorable interface {
	Errorf(string, ...interface{})
}

func disablelog() {
	log.SetLevel(log.FatalLevel)
}

func emptyChain() *Chain {
	disablelog()
	_ = os.RemoveAll(testPath)
	_ = os.Mkdir(testPath, os.ModePerm)
	s := &DiskStore{Folder: testPath}
	c, err := New(s, func([32]byte) bool { return true })
	if err != nil {
		panic(err)
	}
	return c
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
	if !b.Date.Equal(subTime) {
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
			t.Errorf("Broken Linkage! %x has no predecessor", ob.Hash())
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
	disablelog()
	c := dummyChain(t)
	if !c.Valid() {
		t.Error("Validation of valid chain failed")
	}

	b := c.Get(c.LastHash())
	for i := 0; i < 50; i++ {
		b = c.Get(b.PrevHash)
	}
	h := b.Hash()
	fname := path.Join(testPath, base64.URLEncoding.EncodeToString(h[:]))
	err := os.Remove(fname)
	if err != nil {
		t.Error(err)
	}
	file, err := os.Create(fname)
	if err != nil {
		t.Errorf("Could not open file: %s", fname)
	}
	b.Content = "MODIFIED"
	enc := gob.NewEncoder(file)
	err = enc.Encode(*b)
	if err != nil {
		t.Error(err)
	}
	file.Close()
	if c.Valid() {
		t.Error("Modified Chain reported as valid")
	}
}

func TestRestore(t *testing.T) {
	rpath := path.Join(os.TempDir(), "uspeak-core-test-restore")
	_ = os.RemoveAll(rpath)
	_ = os.Mkdir(rpath, os.ModePerm)
	s := &DiskStore{Folder: rpath}
	c1, err := New(s, func([32]byte) bool { return true })
	if err != nil {
		t.Errorf("Chain creation failed: %s", err)
	}
	b1 := Block{PrevHash: c1.LastHash(), Content: "Block", Nonce: 0, Date: time.Now()}
	h := b1.Hash()
	_, err = c1.Add(b1)
	if err != nil {
		t.Error(err)
	}

	c2, err := New(s, func([32]byte) bool { return true })
	if c2.Length() != c1.Length() {
		t.Error("Chains are not the same length")
	}
	b2 := c2.Get(h)
	if !(b2 != nil && b2.Content == "Block") {
		t.Error("Chains did not contain the same block")
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
