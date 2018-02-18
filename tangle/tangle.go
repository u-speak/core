package tangle

import (
	"math/rand"

	"github.com/u-speak/core/post"
	"github.com/u-speak/core/tangle/datastore"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"

	log "github.com/sirupsen/logrus"
)

const (
	// MinimumWeight for new site.Sites
	MinimumWeight = 1
	// MinimumValidations specifies how many sites must be verified by a new site
	MinimumValidations = 2
	// MaxRecommendations specifies how many sites can be returned by RecommendTips
	MaxRecommendations = 4
)

// Tangle stores the relation between different transactions
type Tangle struct {
	tips  map[hash.Hash]bool
	store store.Store
	data  *datastore.Store
}

// Options are used for initial configuration
type Options struct {
	Store    store.Store
	DataPath string
}

// Object is the exposed site including the content
type Object struct {
	Site *site.Site
	Data datastore.Serializable
}

// New returns a fresh initialized tangle
func New(o Options) (*Tangle, error) {
	ds, err := datastore.New(o.DataPath)
	if err != nil {
		return nil, err
	}
	t := &Tangle{data: ds}
	return t, t.Init(o)
}

// Init initializes the tangle with two genesis blocks
func (t *Tangle) Init(o Options) error {
	t.tips = make(map[hash.Hash]bool)
	t.store = o.Store
	if store.Empty(t.store) {
		gen1 := &site.Site{Content: hash.Hash{24, 67, 68, 72, 132, 181}, Nonce: 373, Type: "genesis"}
		gen2 := &site.Site{Content: hash.Hash{24, 67, 68, 72, 132, 182}, Nonce: 510, Type: "genesis"}
		err := t.store.Add(gen1)
		if err != nil {
			return err
		}
		err = t.store.Add(gen2)
		if err != nil {
			return err
		}
		t.store.SetTips(gen1.Hash(), nil)
		t.store.SetTips(gen2.Hash(), nil)
	}
	for _, tip := range t.store.GetTips() {
		t.tips[tip] = true
	}
	return nil
}

// Add Validates the site and adds it to the tangle
// to be valid, a site has to:
// * Validate at least one tip
// * Have a weight of at least MinimumWeight
func (t *Tangle) Add(s *Object) error {
	err := t.verifySite(s.Site)
	if err != nil {
		return err
	}
	v := func() bool {
		for _, v := range s.Site.Validates {
			if t.HasTip(v.Hash()) {
				return true
			}
		}
		return false
	}()
	if !v {
		return ErrNotValidating
	}
	return t.addSite(s, true)
}

// Size returns the amount of sites in the tangle
func (t *Tangle) Size() int {
	return t.store.Size()
}

// Tips returns a list of unconfirmed tips
func (t *Tangle) Tips() []*site.Site {
	keys := []*site.Site{}
	for h := range t.tips {
		s := t.Get(h)
		if s != nil {
			keys = append(keys, s.Site)
		}
	}
	return keys
}

// Get retrieves the specified site
func (t *Tangle) Get(h hash.Hash) *Object {
	md := t.store.Get(h)
	if md == nil {
		return nil
	}
	var data datastore.Serializable
	switch md.Type {
	case "post":
		p := &post.Post{}
		err := t.data.Get(p, md.Content)
		if err != nil {
			log.Error(err)
			return nil
		}
		data = p
	case "genesis":
		data = nil
	case "dummy":
		d := &dummydata{}
		err := t.data.Get(d, md.Content)
		if err != nil {
			log.Error(err)
			return nil
		}
		data = d
	default:
		log.Errorf("Type `%s' not implemented", md.Type)
		return nil
	}
	return &Object{Site: md, Data: data}
}

// Close closes the underlying store
func (t *Tangle) Close() {
	t.store.Close()
	t.data.Close()
}

// HasTip checks if the specified hash is a tip of the current tangle
func (t *Tangle) HasTip(h hash.Hash) bool {
	return t.tips[h]
}

// Weight returns the weight of a specific site inside the tangle
func (t *Tangle) Weight(s *site.Site) int {
	bound := make(map[*site.Site]bool)
	// Setting up exclusion list
	excl := make(map[hash.Hash]bool)

	inject := func(l []*site.Site) {
		for _, v := range l {
			bound[v] = true
		}
	}

	inject(s.Validates)
	for len(bound) != 0 {
		for st := range bound {
			delete(bound, st)
			excl[st.Hash()] = true
			for _, v := range st.Validates {
				if !excl[v.Hash()] {
					bound[v] = true
				}
			}
		}
	}
	// Calculating weight
	rvl := make(map[hash.Hash][]*site.Site)
	inject(t.Tips())
	for len(bound) != 0 {
		for st := range bound {
			delete(bound, st)
			for _, v := range st.Validates {
				if !excl[v.Hash()] {
					bound[v] = true
					rvl[v.Hash()] = append(rvl[v.Hash()], st)
				}
			}
		}
	}
	w := s.Hash().Weight()
	inject(rvl[s.Hash()])
	for len(bound) != 0 {
		for st := range bound {
			delete(bound, st)
			excl[st.Hash()] = true
			w += st.Hash().Weight()
			for _, v := range rvl[st.Hash()] {
				if !excl[v.Hash()] {
					bound[v] = true
				}
			}
		}
	}
	return w
}

// Hashes returns all stored hashes
func (t *Tangle) Hashes() []hash.Hash {
	return t.store.Hashes()
}

// RecommendTips returns tips to be used
func (t *Tangle) RecommendTips() []*site.Site {
	recs := t.Tips()
	if len(recs) > MinimumValidations {
		return recs[:MaxRecommendations]
	}
	blst := make(map[hash.Hash]bool)
	hashes := t.Hashes()
	for _, tip := range recs {
		blst[tip.Hash()] = true
	}
	for len(recs) < MinimumValidations {
		rndhash := hashes[rand.Int()%len(hashes)]
		if blst[rndhash] {
			continue
		}
		blst[rndhash] = true
		s := t.Get(rndhash)
		if s == nil {
			continue
		}
		recs = append(recs, s.Site)
	}
	return recs
}

// Inject adds sites to the tangle without checking for validated tips
func (t *Tangle) Inject(s *Object, tip bool) error {
	err := t.verifySite(s.Site)
	if err != nil {
		return err
	}
	return t.addSite(s, tip)
}

func (t *Tangle) verifySite(s *site.Site) error {
	if s.Hash().Weight() < MinimumWeight {
		return ErrWeightTooLow
	}
	if len(s.Validates) < MinimumValidations {
		return ErrTooFewValidations
	}
	return nil
}

func (t *Tangle) addSite(s *Object, tip bool) error {
	for _, vs := range s.Site.Validates {
		delete(t.tips, vs.Hash())
	}
	if tip {
		t.tips[s.Site.Hash()] = true
		t.store.SetTips(s.Site.Hash(), s.Site.Validates)
	}

	err := t.store.Add(s.Site)
	if err != nil {
		return err
	}
	err = t.data.Put(s.Data)
	if err != nil {
		return err
	}
	return nil
}
