package boltstore

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	s := BoltStore{}
	err := s.Init(store.Options{Path: "/tmp/testInit.db"})
	assert.NoError(t, err)
	s.Close()
	os.Remove("/tmp/testInit.db")
}

func TestTips(t *testing.T) {
	s := BoltStore{}
	err := s.Init(store.Options{Path: "/tmp/testTips.db"})
	assert.NoError(t, err)
	defer os.Remove("/tmp/testTips.db")
	assert.Empty(t, s.GetTips())
	tips := []hash.Hash{hash.Hash{1}, hash.Hash{2}, hash.Hash{3}}
	s.SetTips(tips, []hash.Hash{})
	assert.Equal(t, tips, s.GetTips())
	s.SetTips([]hash.Hash{hash.Hash{4}}, []hash.Hash{hash.Hash{2}})
	assert.Len(t, s.GetTips(), 3)
	assert.Contains(t, s.GetTips(), hash.Hash{4})
	assert.NotContains(t, s.GetTips(), hash.Hash{2})
}

func TestAddGet(t *testing.T) {
	s := BoltStore{}
	err := s.Init(store.Options{Path: "/tmp/testAddGet.db"})
	assert.NoError(t, err)
	defer s.Close()
	defer os.Remove("/tmp/testAddGet.db")

	site1 := &site.Site{Content: hash.Hash{1, 3, 3, 7}}
	site2 := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Validates: []*site.Site{site1}}
	site3 := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Validates: []*site.Site{site1, site2}}
	gs := &site.Site{}

	err = s.Add(site1)
	assert.NoError(t, err)
	err = gs.Deserialize(s.Get(site1.Hash()))
	assert.NoError(t, err)
	assert.Equal(t, site1, gs)

	err = s.Add(site2)
	assert.NoError(t, err)
	err = gs.Deserialize(s.Get(site2.Hash()))
	assert.NoError(t, err)
	assert.Equal(t, site2, gs)

	err = s.Add(site3)
	assert.NoError(t, err)
	err = gs.Deserialize(s.Get(site3.Hash()))
	assert.NoError(t, err)
	assert.Equal(t, site3, gs)
	assert.Equal(t, site2, gs.Validates[1])
}
