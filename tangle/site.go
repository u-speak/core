package tangle

import (
	"bytes"
	"golang.org/x/crypto/blake2s"
	"text/template"

	log "github.com/sirupsen/logrus"
)

// Site represents a single storage node inside the tangle
type Site struct {
	Validates []*Site
	Nonce     uint64
	Content   Hash
}

// Hash computes the hash of the site
func (s *Site) Hash() Hash {
	t := template.Must(template.New("site").Parse("C{{.Content}}N{{.Nonce}}{{range $k,$s := .Validates}}V{{$s.Hash}}{{end}}"))
	w := &bytes.Buffer{}
	err := t.Execute(w, s)
	if err != nil {
		log.Error(err)
	}
	return blake2s.Sum256(w.Bytes())
}

func (s *Site) mine(targetWeight int) {
	for s.Hash().Weight() < targetWeight {
		s.Nonce++
	}
}
