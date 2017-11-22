package chain

import (
	"encoding/base64"
	"encoding/gob"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

var testPath = path.Join(os.TempDir(), "uspeak-core-test")

func emptyDiskStore() *DiskStore {
	disablelog()
	_ = os.RemoveAll(testPath)
	_ = os.Mkdir(testPath, os.ModePerm)
	s := &DiskStore{Folder: testPath}
	return s
}

func TestDiskStoreInitialization(t *testing.T) {
	_ = os.RemoveAll(testPath)
	_ = os.Mkdir(testPath, os.ModePerm)
	s := &DiskStore{Folder: testPath}
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

func TestDiskStoreAddGet(t *testing.T) {
	s := emptyDiskStore()
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
	}
	if !b.Date.Equal(subTime) {
		t.Errorf("Wrong submission time. Expected %s, got %s", subTime.String(), b.Date.String())
	}
	if b.Content != "test" {
		t.Errorf("Wrong content. Expected test, got %s", b.Content)
	}
}

func TestDiskStoreValidation(t *testing.T) {
	disablelog()
	c := dummyChain(t, emptyDiskStore())
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

func TestDiskStoreRestore(t *testing.T) {
	rpath := path.Join(os.TempDir(), "uspeak-core-test-restore")
	_ = os.RemoveAll(rpath)
	_ = os.Mkdir(rpath, os.ModePerm)
	s := &DiskStore{Folder: rpath}
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

	c2, err := New(s, func(Hash) bool { return true })
	if c2.Length() != c1.Length() {
		t.Error("Chains are not the same length")
	}
	b2 := c2.Get(h)
	if !(b2 != nil && b2.Content == "Block") {
		t.Error("Chains did not contain the same block")
	}
}

func BenchmarkDiskStoreAdd(b *testing.B) {
	s := emptyDiskStore()
	s.Init()
	for i := 0; i < b.N; i++ {
		err := s.Add(Block{Content: "Block" + strconv.Itoa(i), Nonce: 0, Date: time.Now()})
		if err != nil {
			b.Errorf("Error adding Block: %s", err)
		}
	}
}

func keys(b *DiskStore) []Hash {
	hkeys := []Hash{}
	files, err := ioutil.ReadDir(b.Folder)
	if err != nil {
		return nil
	}
	for _, f := range files {
		stat := Hash{}
		h, err := base64.URLEncoding.DecodeString(f.Name())
		if err != nil {
			continue
		}
		copy(stat[:], h)
		bl := b.Get(stat)
		if bl != nil {
			hkeys = append(hkeys, stat)
		}
	}
	return hkeys
}

func BenchmarkDiskStoreGet(b *testing.B) {
	c := dummyChain(b, emptyDiskStore())
	ks := keys(c.blocks.(*DiskStore))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := rand.Intn(len(ks))
		c.Get(ks[k])
	}
}

func BenchmarkDiskStoreValidation(b *testing.B) {
	c := dummyChain(b, emptyDiskStore())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !c.Valid() {
			b.Error("Chain validation failed")
		}
	}
}
