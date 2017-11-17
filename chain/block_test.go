package chain

import (
	"testing"
	"time"
)

func TestHash(t *testing.T) {
	b := Block{}
	baseHash := b.Hash()

	cb := b
	cb.Content = "Foo"
	if cb.Hash() == baseHash {
		t.Error("Hash did not change while modifying Content")
	}

	db := b
	db.Date = time.Now()
	if db.Hash() == baseHash {
		t.Error("Hash did not change while modifying Date")
	}

	nb := b
	nb.Nonce = 42
	if db.Hash() == baseHash {
		t.Error("Hash did not change while modifying Nonce")
	}

	pb := b
	pb.PrevHash = [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	if pb.Hash() == baseHash {
		t.Error("Hash did not change while modifying Previous Hash")
	}

	pkb := b
	pkb.PubKey = "--- BEGIN PGP KEY --- ..... --- END PGP KEY ---"
	if pkb.Hash() == baseHash {
		t.Error("Hash did not change while modifying Public Key")
	}

	sb := b
	sb.Signature = "--- BEGIN PGP SIGNATURE --- ..... --- END PGP SIGNATURE ---"
	if sb.Hash() == baseHash {
		t.Error("Hash did not change while modifying Signature")
	}

	tb := b
	tb.Type = "test"
	if tb.Hash() == baseHash {
		t.Error("Hash did not change while modifying Type")
	}
}

func TestSerialization(t *testing.T) {
	testTime := time.Now()
	blueprint := &Block{
		Nonce:     42,
		Content:   "foo",
		Signature: "bar",
		PubKey:    "baz",
		PrevHash:  [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		Type:      "post",
		Date:      testTime,
	}
	enc, err := blueprint.encode()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	dec, err := DecodeBlock(enc)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	switch {
	case dec.Nonce != blueprint.Nonce:
		t.Errorf("Nonce was modified during encoding! %d -> %d", blueprint.Nonce, dec.Nonce)
	case dec.Content != blueprint.Content:
		t.Errorf("Content was modified during encoding! %s -> %s", blueprint.Content, dec.Content)
	case dec.Signature != blueprint.Signature:
		t.Errorf("Signature was modified during encoding! %s -> %s", blueprint.Signature, dec.Signature)
	case dec.PubKey != blueprint.PubKey:
		t.Errorf("PubKey was modified during encoding! %s -> %s", blueprint.PubKey, dec.PubKey)
	case dec.PrevHash != blueprint.PrevHash:
		t.Errorf("PrevHash was modified during encoding! %+v -> %+v", blueprint.PrevHash, dec.PrevHash)
	case dec.Type != blueprint.Type:
		t.Errorf("Type was modified during encoding! %s -> %s", blueprint.Type, dec.Type)
	case !dec.Date.Equal(blueprint.Date):
		t.Error("Dates did not match")
	}
}

func BenchmarkDecode(b *testing.B) {
	bl := &Block{
		Nonce:     42,
		Content:   "foo",
		Signature: "bar",
		PubKey:    "baz",
		PrevHash:  [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
		Type:      "post",
		Date:      time.Now(),
	}
	enc, err := bl.encode()
	if err != nil {
		b.Error(b)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeBlock(enc)
	}
}
