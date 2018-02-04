package tangle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init()
	assert.NoError(t, err)
	assert.Equal(t, 2, tngl.Size())
}

func TestTips(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init()
	assert.NoError(t, err)
	assert.Len(t, tngl.Tips(), 2)
}

func TestGet(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init()
	assert.NoError(t, err)
	assert.Nil(t, tngl.Get(Hash{}))
	for _, s := range tngl.Tips() {
		assert.Equal(t, s, tngl.Get(s.Hash()))
	}
}

func TestAdd(t *testing.T) {
	tngl := Tangle{}
	err := tngl.Init()
	assert.NoError(t, err)
	tips := tngl.Tips()
	err = tngl.Add(&Site{Content: Hash{1, 3, 3, 7}, Nonce: 0})
	assert.Equal(t, ErrWeightTooLow, err)
	err = tngl.Add(&Site{Content: Hash{1, 3, 3, 7}, Nonce: 263})
	assert.Equal(t, ErrTooFewValidations, err)

	sub := &Site{Content: Hash{1, 3, 3, 7}, Nonce: 0, Validates: []*Site{tips[0], tips[1]}}
	sub.mine(1)
	t.Logf("Finished mining with nonce: %d", sub.Nonce)
	err = tngl.Add(sub)
	assert.NoError(t, err)
	assert.False(t, tngl.tips[tips[0]])
	assert.False(t, tngl.tips[tips[1]])
	assert.True(t, tngl.tips[sub])
	assert.Equal(t, sub, tngl.Get(sub.Hash()))
}
