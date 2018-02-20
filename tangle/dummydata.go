package tangle

import (
	"github.com/u-speak/core/tangle/hash"
)

type dummydata struct {
	content string
}

func (d *dummydata) Hash() (hash.Hash, error) {
	return hash.New([]byte(d.content)), nil
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

func (d *dummydata) JSON() error {
	return nil
}

func (d *dummydata) ReInit() error {
	return nil
}
