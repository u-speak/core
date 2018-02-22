package node

import (
	"github.com/u-speak/core/tangle"
)

// FromObject converts a regular site into a distribution ready site
func FromObject(o *tangle.Object) (*Site, error) {
	vs := [][]byte{}
	for _, v := range o.Site.Validates {
		vs = append(vs, v.Hash().Slice())
	}
	data, err := o.Data.Serialize()
	if err != nil {
		return nil, err
	}
	return &Site{
		Validates: vs,
		Nonce:     o.Site.Nonce,
		Content:   o.Site.Content.Slice(),
		Type:      o.Site.Type,
		Data:      data,
	}, nil
}
