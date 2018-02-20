package tangle

import (
	"github.com/u-speak/core/tangle/hash"
)

type genesis struct {
	Content string `json:"content"`
}

func (d *genesis) Hash() (hash.Hash, error) {
	return hash.Hash{24, 67, 68, 72, 132}, nil
}
func (d *genesis) Serialize() ([]byte, error) {
	return []byte("GENESIS"), nil
}
func (d *genesis) Deserialize(bts []byte) error {
	return nil
}
func (d *genesis) Type() string {
	return "genesis"
}

func (d *genesis) JSON() error {
	d.Content = "GENESIS"
	return nil
}

func (d *genesis) ReInit() error {
	return nil
}
