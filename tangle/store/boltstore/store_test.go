package boltstore

import (
	"github.com/u-speak/core/tangle"
	"github.com/u-speak/core/tangle/hash"
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

func TestAddGet(t *testing.T) {
	s := BoltStore{}
	err := s.Init(store.Options{Path: "/tmp/testAddGet.db"})
	assert.NoError(t, err)
	defer s.Close()
	defer os.Remove("/tmp/testAddGet.db")

	site1 := &tangle.Site{Content: hash.Hash{1, 3, 3, 7}}
	site2 := &tangle.Site{Content: hash.Hash{1, 3, 3, 7}, Validates: []*tangle.Site{site1}}
	site3 := &tangle.Site{Content: hash.Hash{1, 3, 3, 7}, Validates: []*tangle.Site{site1, site2}}
	gs := &tangle.Site{}

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
