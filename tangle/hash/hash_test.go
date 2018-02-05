package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeight(t *testing.T) {
	assert.Equal(t, 0, Hash{1, 3, 3, 7}.Weight())
	assert.Equal(t, 1, Hash{0, 3, 3, 7}.Weight())
	assert.Equal(t, 1, Hash{0, 3, 0, 7}.Weight())
	assert.Equal(t, 5, Hash{0, 0, 0, 0, 0, 5}.Weight())
}

func TestSlice(t *testing.T) {
	assert.Equal(t, []byte{1, 3, 3, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Hash{1, 3, 3, 7}.Slice())
}
