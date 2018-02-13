package tangle

import (
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
)

// Tangle stores the relation between different transactions
type Tangle struct {
	tips  map[*site.Site]bool
	sites map[hash.Hash]*site.Site
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
	t.tips = make(map[*site.Site]bool)
	t.sites = make(map[hash.Hash]*site.Site)
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
		t.store.SetTips(gen1, nil)
		t.store.SetTips(gen2, nil)
	}
	for _, tip := range t.store.GetTips() {
		t.tips[t.store.Get(tip)] = true
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
			if t.hasTip(v) {
				return true
			}
		}
		return false
	}()
	if !v {
		return ErrNotValidating
	}
	for _, vs := range s.Site.Validates {
		delete(t.tips, vs)
	}
	t.tips[s.Site] = true
	t.store.SetTips(s.Site, s.Site.Validates)
	return t.addSite(s)
}

// Size returns the amount of sites in the tangle
func (t *Tangle) Size() int {
	return t.store.Size()
}

// Tips returns a list of unconfirmed tips
func (t *Tangle) Tips() []*site.Site {
	keys := []*site.Site{}
	for s := range t.tips {
		keys = append(keys, s)
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
	}
	return &Object{Site: md, Data: data}
}

// Close closes the underlying store
func (t *Tangle) Close() {
	t.store.Close()
	t.data.Close()
}

func (t *Tangle) hasTip(s *site.Site) bool {
	return t.tips[s]
}

func (t *Tangle) weight(s *site.Site) int {
	bound := make(map[*site.Site]bool)
	// Setting up exclusion list
	excl := make(map[*site.Site]bool)

	inject := func(l []*site.Site) {
		for _, v := range l {
			bound[v] = true
		}
	}

	inject(s.Validates)
	for len(bound) != 0 {
		for st := range bound {
			delete(bound, st)
			excl[st] = true
			for _, v := range st.Validates {
				if !excl[v] {
					bound[v] = true
				}
			}
		}
	}
	// Calculating weight
	rvl := make(map[*site.Site][]*site.Site)
	inject(t.Tips())
	for len(bound) != 0 {
		for st := range bound {
			delete(bound, st)
			for _, v := range st.Validates {
				if !excl[v] {
					bound[v] = true
					rvl[v] = append(rvl[v], st)
				}
			}
		}
	}
	w := s.Hash().Weight()
	inject(rvl[s])
	for len(bound) != 0 {
		for st := range bound {
			delete(bound, st)
			excl[st] = true
			w += st.Hash().Weight()
			for _, v := range rvl[st] {
				if !excl[v] {
					bound[v] = true
				}
			}
		}
	}
	return w
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

func (t *Tangle) addSite(s *Object) error {
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
