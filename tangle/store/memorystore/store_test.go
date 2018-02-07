package memorystore

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	s := MemoryStore{}
	err := s.Init(store.Options{})
	assert.NoError(t, err)
	s.Close()
}

func TestTips(t *testing.T) {
	s := MemoryStore{}
	err := s.Init(store.Options{})
	assert.NoError(t, err)
	assert.Empty(t, s.GetTips())
	s1 := &site.Site{Content: hash.Hash{1}}
	s2 := &site.Site{Content: hash.Hash{2}}
	s3 := &site.Site{Content: hash.Hash{3}}
	s.SetTips(s1, nil)
	assert.Equal(t, []hash.Hash{s1.Hash()}, s.GetTips())
	s.SetTips(s2, nil)
	assert.Equal(t, []hash.Hash{s1.Hash(), s2.Hash()}, s.GetTips())
	s.SetTips(s3, []*site.Site{s2})
	assert.Len(t, s.GetTips(), 2)
	assert.NotContains(t, s.GetTips(), hash.Hash{2})
}

func TestAddGet(t *testing.T) {
	s := MemoryStore{}
	err := s.Init(store.Options{})
	assert.NoError(t, err)
	defer s.Close()

	site1 := &site.Site{Content: hash.Hash{1, 3, 3, 7}}
	site2 := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Validates: []*site.Site{site1}}
	site3 := &site.Site{Content: hash.Hash{1, 3, 3, 7}, Validates: []*site.Site{site1, site2}}

	err = s.Add(site1)
	assert.NoError(t, err)
	assert.Equal(t, site1, s.Get(site1.Hash()))

	err = s.Add(site2)
	assert.NoError(t, err)
	assert.Equal(t, site2, s.Get(site2.Hash()))

	err = s.Add(site3)
	assert.NoError(t, err)
	assert.Equal(t, site3, s.Get(site3.Hash()))
	assert.Equal(t, site2, s.Get(site3.Hash()).Validates[1])
}
