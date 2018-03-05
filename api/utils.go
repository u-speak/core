package api

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/u-speak/core/post"
	"github.com/u-speak/core/tangle"
	"github.com/u-speak/core/tangle/datastore"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/util"
)

// JSONize converts an object into a jsonSite
func JSONize(o *tangle.Object) jsonSite {
	h := o.Site.Hash()
	vals := []string{}
	for _, v := range o.Site.Validates {
		vals = append(vals, v.Hash().String())
	}
	return jsonSite{
		Nonce:        o.Site.Nonce,
		Hash:         h.String(),
		Validates:    vals,
		Content:      o.Site.Content.String(),
		Type:         o.Site.Type,
		BubbleBabble: util.EncodeBubbleBabble(h),
		Data:         o.Data,
	}
}

func decodeImageHash(s string) (hash.Hash, string) {
	a := strings.Split(s, ".")
	h, _ := DecodeHash(a[0])
	if len(a) == 1 {
		return h, ""
	}
	switch a[1] {
	case "png":
		return h, "image/png"
	case "jpg", "jpeg":
		return h, "image/jpeg"
	}
	return h, ""
}

// DecodeHash is a utility function, allowing the decoding of various formats
func DecodeHash(s string) (hash.Hash, error) {
	h := [32]byte{}
	var hs []byte
	h, err := util.DecodeBubbleBabble(s)
	if err == nil {
		return h, nil
	}
	hs, err = base64.URLEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	hs, err = base64.StdEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	hs, err = base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	hs, err = base64.RawStdEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	return [32]byte{}, errors.New("Could not parse base64 data")
}

func verifyGPG(s datastore.Serializable) error {
	err := s.ReInit()
	if err != nil {
		return err
	}
	_, err = s.(*post.Post).Verify()
	return err
}
