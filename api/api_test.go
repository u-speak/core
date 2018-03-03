package api

import (
	"encoding/base64"
	"testing"
)

var validHash = [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

const invalid = "InVaLiDsTrInG"

func TestDecodeHash(t *testing.T) {
	strings := []string{
		base64.URLEncoding.EncodeToString(validHash[:]),
		base64.RawURLEncoding.EncodeToString(validHash[:]),
		base64.StdEncoding.EncodeToString(validHash[:]),
		base64.RawStdEncoding.EncodeToString(validHash[:]),
	}
	for _, s := range strings {
		h, err := DecodeHash(s)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if h != validHash {
			t.Errorf("Error decoding hash! Expected %v, got %v", h, validHash)
		}
	}
	_, err := DecodeHash(invalid)
	if err == nil {
		t.Error("Expected error but got none")
	}

}

func TestDecodeImageHash(t *testing.T) {
	cases := map[string]string{
		base64.URLEncoding.EncodeToString(validHash[:]) + ".png":     "image/png",
		base64.URLEncoding.EncodeToString(validHash[:]) + ".jpg":     "image/jpeg",
		base64.URLEncoding.EncodeToString(validHash[:]) + ".jpeg":    "image/jpeg",
		base64.URLEncoding.EncodeToString(validHash[:]) + ".schmafu": "",
	}
	for i, o := range cases {
		_, ty := decodeImageHash(i)
		if ty != o {
			t.Errorf("Wrong imagetype! Expected: %v, got: %v", o, ty)
		}
	}
}
