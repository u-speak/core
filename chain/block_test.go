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
