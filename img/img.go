package img

import (
	"bytes"
	"encoding/base64"
	"image"
	// Used for decoding
	_ "image/gif"
	// Used for decoding
	_ "image/jpeg"
	// Used for decoding
	_ "image/png"

	"github.com/u-speak/core/tangle/hash"
)

// Image wraps the raw byte data of the image
type Image struct {
	Raw []byte
}

// Hash returns the hash for storage
func (i *Image) Hash() (hash.Hash, error) {
	b := base64.URLEncoding.EncodeToString(i.Raw)
	return hash.New([]byte(b)), nil
}

// Serialize implements tangle/datastore.serializable
func (i *Image) Serialize() ([]byte, error) {
	return i.Raw, nil
}

// Deserialize implements tangle/datastore.serializable
func (i *Image) Deserialize(bts []byte) error {
	i.Raw = bts
	return nil
}

// ReInit restores the original field after serialization
func (i *Image) ReInit() error { return nil }

// JSON prepares for json encoding
func (i *Image) JSON() error { return nil }

// Type implements tangle/datastore.serializable
func (i *Image) Type() string { return "image" }

// Image returns the native golang image representation
func (i *Image) Image() (image.Image, error) {
	buff := bytes.NewBuffer(i.Raw)
	img, _, err := image.Decode(buff)
	return img, err
}
