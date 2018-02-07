package site

import (
	"bytes"
	"golang.org/x/crypto/blake2s"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/tangle/hash"
	"github.com/vmihailenco/msgpack"
)

// Site represents a single storage node inside the tangle
type Site struct {
	Validates []*Site
	Nonce     uint64
	Content   hash.Hash
}

// Hash computes the hash of the site
func (s *Site) Hash() hash.Hash {
	t := template.Must(template.New("site").Parse("C{{.Content}}N{{.Nonce}}{{range $k,$s := .Validates}}V{{$s.Hash}}{{end}}"))
	w := &bytes.Buffer{}
	err := t.Execute(w, s)
	if err != nil {
		log.Error(err)
	}
	return blake2s.Sum256(w.Bytes())
}

// Serialize converts the site to a slice of bytes
func (s *Site) Serialize() []byte {
	b, _ := msgpack.Marshal(s)
	return b
}

// Deserialize restores the site from a slice of bytes
func (s *Site) Deserialize(b []byte) error {
	return msgpack.Unmarshal(b, s)
}

func (s *Site) mine(targetWeight int) {
	for s.Hash().Weight() < targetWeight {
		s.Nonce++
	}
}
