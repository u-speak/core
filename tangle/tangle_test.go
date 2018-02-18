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

func dd(s string) *dummydata {
	return &dummydata{content: s}
}

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
	tngl, err := New(Options{Store: ms(), DataPath: path.Join(os.TempDir(), "testget")})
	assert.NoError(t, err)
	assert.Nil(t, tngl.Get(hash.Hash{}))
	for _, s := range tngl.Tips() {
		assert.Equal(t, s, tngl.Get(s.Hash()).Site)
	}
}

func TestAdd(t *testing.T) {
	tngl, err := New(Options{Store: ms(), DataPath: path.Join(os.TempDir(), "testadd")})
	assert.NoError(t, err)
	tips := tngl.Tips()
	err = tngl.Add(&Object{Site: &site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0}, Data: dd("1337")})
	assert.Equal(t, ErrWeightTooLow, err)
	st := &Object{Site: &site.Site{Content: hash.Hash{1, 3, 3, 7}}, Data: dd("1337")}
	st.Site.Mine(1)
	err = tngl.Add(st)
	assert.Equal(t, ErrTooFewValidations, err)

	h, _ := dd("1337").Hash()
	sub := &Object{Site: &site.Site{Content: h, Nonce: 0, Validates: []*site.Site{tips[0], tips[1]}, Type: "dummy"}, Data: dd("1337")}
	sub.Site.Mine(1)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	assert.False(t, tngl.tips[tips[0].Hash()])
	assert.False(t, tngl.tips[tips[1].Hash()])
	assert.True(t, tngl.tips[sub.Site.Hash()])
	assert.Equal(t, sub, tngl.Get(sub.Site.Hash()))
}

func TestRestore(t *testing.T) {
	dbpath := path.Join(os.TempDir(), "testRestore.db")
	defer os.Remove(dbpath)
	datapath := path.Join(os.TempDir(), "testRestoreData.db")
	defer os.Remove(datapath)
	bs := boltstore.BoltStore{}
	err := bs.Init(store.Options{Path: dbpath})
	assert.NoError(t, err)
	tngl, err := New(Options{Store: &bs, DataPath: datapath})
	assert.NoError(t, err)
	tips := tngl.Tips()
	sub := &Object{Site: &site.Site{Content: hash.Hash{1, 3, 3, 7}, Nonce: 0, Validates: []*site.Site{tips[0], tips[1]}, Type: "dummy"}, Data: dd("1337")}
	sub.Site.Mine(1)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	tips = tngl.Tips()
	tngl.Close()

	bs2 := boltstore.BoltStore{}
	err = bs2.Init(store.Options{Path: dbpath})
	assert.NoError(t, err)
	tngl2, err := New(Options{Store: &bs2, DataPath: datapath})
	assert.NoError(t, err)
	assert.Equal(t, tips, tngl2.Tips())
}

func TestWeight(t *testing.T) {
	dbpath := path.Join(os.TempDir(), "testweight.db")
	defer os.Remove(dbpath)
	tngl, err := New(Options{Store: ms(), DataPath: dbpath})
	assert.NoError(t, err)
	tips := tngl.Tips()
	gen1, gen2 := tips[0], tips[1]
	s1 := &Object{Site: &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 53}, Nonce: 0, Validates: []*site.Site{gen1, gen2}}, Data: dd("s1")}
	s2 := &Object{Site: &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 54}, Nonce: 0, Validates: []*site.Site{s1.Site, gen2}}, Data: dd("s2")}
	s3 := &Object{Site: &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 55}, Nonce: 0, Validates: []*site.Site{s2.Site, s1.Site}}, Data: dd("s3")}
	s4 := &Object{Site: &site.Site{Content: hash.Hash{72, 132, 196, 211, 77, 56}, Nonce: 0, Validates: []*site.Site{s3.Site, s2.Site}}, Data: dd("s4")}
	s1.Site.Mine(1)
	s2.Site.Mine(1)
	s3.Site.Mine(1)
	s4.Site.Mine(1)
	assert.NoError(t, tngl.Add(s1))
	assert.NoError(t, tngl.Add(s2))
	assert.NoError(t, tngl.Add(s3))
	assert.NoError(t, tngl.Add(s4))
	assert.EqualValues(t, 6, tngl.Size())
	tngl.Weight(s2.Site)
	assert.EqualValues(t, s4.Site.Hash().Weight(), tngl.Weight(s4.Site))
	assert.EqualValues(t, s4.Site.Hash().Weight()+s3.Site.Hash().Weight(), tngl.Weight(s3.Site))
	assert.EqualValues(t, s4.Site.Hash().Weight()+s3.Site.Hash().Weight()+s2.Site.Hash().Weight(), tngl.Weight(s2.Site))
	assert.EqualValues(t, s4.Site.Hash().Weight()+s3.Site.Hash().Weight()+s2.Site.Hash().Weight()+s1.Site.Hash().Weight(), tngl.Weight(s1.Site))
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
	assert.NoError(b, tngl.Add(&Object{Site: s1, Data: dd("s1")}))
	assert.NoError(b, tngl.Add(&Object{Site: s2, Data: dd("s2")}))
	assert.NoError(b, tngl.Add(&Object{Site: s3, Data: dd("s3")}))
	assert.NoError(b, tngl.Add(&Object{Site: s4, Data: dd("s4")}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tngl.Weight(s1)
	}
}
