package tangle

import (
	"github.com/u-speak/core/tangle/hash"

	"golang.org/x/crypto/blake2s"
)

type dummydata struct {
	content string
}

func (d *dummydata) Hash() (hash.Hash, error) {
	return blake2s.Sum256([]byte(d.content)), nil
}
func (d *dummydata) Serialize() ([]byte, error) {
	return []byte(d.content), nil
}
func (d *dummydata) Deserialize(bts []byte) error {
	d.content = string(bts)
	return nil
}
func (d *dummydata) Type() string {
	return "dummy"
}
