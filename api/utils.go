package api

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/util"
)

func jsonize(b *chain.Block) jsonBlock {
	h := b.Hash()
	p := b.PrevHash
	return jsonBlock{
		Hash:         base64.URLEncoding.EncodeToString(h[:]),
		Nonce:        b.Nonce,
		PrevHash:     base64.URLEncoding.EncodeToString(p[:]),
		Content:      b.Content,
		Signature:    b.Signature,
		Type:         b.Type,
		PubKey:       b.PubKey,
		Date:         b.Date.Unix(),
		BubbleBabble: util.EncodeBubbleBabble(h),
	}
}

func decodeImageHash(s string) (chain.Hash, string) {
	a := strings.Split(s, ".")
	h, _ := decodeHash(a[0])
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

func decodeHash(s string) (chain.Hash, error) {
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
