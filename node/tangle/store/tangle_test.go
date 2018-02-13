package tangle

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"
	"github.com/u-speak/core/tangle/store/boltstore"
	"github.com/u-speak/core/tangle/store/memorystore"
)

func ms() *memorystore.MemoryStore {
	ms := memorystore.MemoryStore{}
	_ = ms.Init(store.Options{})
	return &ms
}

func TestInit(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	assert.Equal(t, 2, tngl.Size())
}

func TestTips(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	assert.Len(t, tngl.Tips(), 2)
}

func TestGet(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	assert.Nil(t, tngl.Get(hash.Hash{}))
	for _, s := range tngl.Tips() {
		assert.Equal(t, s, tngl.Get(s.Hash()))
	}
}

func TestAdd(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	tips := tngl.Tips()
	err = tngl.Add(&site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0})
	assert.Equal(t, ErrWeightTooLow, err)
	err = tngl.Add(&site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 263})
	assert.Equal(t, ErrTooFewValidations, err)

	sub := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0, Validates: []*site.Site{tips[0], tips[1]}}
	sub.Mine(1)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	assert.False(t, tngl.tips[tips[0]])
	assert.False(t, tngl.tips[tips[1]])
	assert.True(t, tngl.tips[sub])
	assert.Equal(t, sub, tngl.Get(sub.Hash()))
}

func TestRestore(t *testing.T) {
	tngl := Tangle{}
	bs := boltstore.BoltStore{}
	err := bs.Init(store.Options{Path: path.Join(os.TempDir(), "testRestore.db")})
	assert.NoError(t, err)
	err = tngl.Init(Options{Store: &bs})
	assert.NoError(t, err)
	tips := tngl.Tips()
	sub := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0, Validates: []*site.Site{tips[0], tips[1]}}
	sub.Mine(1)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	tips = tngl.Tips()
	tngl.Close()

	bs2 := boltstore.BoltStore{}
	err = bs2.Init(store.Options{Path: path.Join(os.TempDir(), "testRestore.db")})
	assert.NoError(t, err)
	tngl2 := Tangle{}
	err = tngl2.Init(Options{Store: &bs2})
	assert.Equal(t, tips, tngl2.Tips())
	os.Remove(path.Join(os.TempDir(), "testRestore.db"))
}

func TestWeight(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(t, err)
	tips := tngl.Tips()
	gen1, gen2 := tips[0], tips[1]
	s1 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 53}, Nonce: 0, Validates: []*site.Site{gen1, gen2}}
	s2 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 54}, Nonce: 0, Validates: []*site.Site{s1, gen2}}
	s3 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 55}, Nonce: 0, Validates: []*site.Site{s2, s1}}
	s4 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 56}, Nonce: 0, Validates: []*site.Site{s3, s2}}
	s1.Mine(1)
	s2.Mine(1)
	s3.Mine(1)
	s4.Mine(1)
	assert.NoError(t, tngl.Add(s1))
	assert.NoError(t, tngl.Add(s2))
	assert.NoError(t, tngl.Add(s3))
	assert.NoError(t, tngl.Add(s4))
	assert.EqualValues(t, 6, tngl.Size())
	tngl.weight(s2)
	assert.EqualValues(t, s4.Hash().Weight(), tngl.weight(s4))
	assert.EqualValues(t, s4.Hash().Weight()+s3.Hash().Weight(), tngl.weight(s3))
	assert.EqualValues(t, s4.Hash().Weight()+s3.Hash().Weight()+s2.Hash().Weight(), tngl.weight(s2))
	assert.EqualValues(t, s4.Hash().Weight()+s3.Hash().Weight()+s2.Hash().Weight()+s1.Hash().Weight(), tngl.weight(s1))
}

func BenchmarkWeight(b *testing.B) {
	tngl := Tangle{}
	err := tngl.Init(Options{Store: ms()})
	assert.NoError(b, err)
	tips := tngl.Tips()
	gen1, gen2 := tips[0], tips[1]
	s1 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 53}, Nonce: 0, Validates: []*site.Site{gen1, gen2}}
	s2 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 54}, Nonce: 0, Validates: []*site.Site{s1, gen2}}
	s3 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 55}, Nonce: 0, Validates: []*site.Site{s2, s1}}
	s4 := &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 56}, Nonce: 0, Validates: []*site.Site{s3, s2}}
	s1.Mine(1)
	s2.Mine(1)
	s3.Mine(1)
	s4.Mine(1)
	assert.NoError(b, tngl.Add(s1))
	assert.NoError(b, tngl.Add(s2))
	assert.NoError(b, tngl.Add(s3))
	assert.NoError(b, tngl.Add(s4))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tngl.weight(s1)
	}
}
