package tangle

import (
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
}

// Options are used for initial configuration
type Options struct {
	Store store.Store
}

// Init initializes the tangle with two genesis blocks
func (t *Tangle) Init(o Options) error {
	t.tips = make(map[*site.Site]bool)
	t.sites = make(map[hash.Hash]*site.Site)
	t.store = o.Store
	if store.Empty(t.store) {
		log.Info("Initializing new Tangle")
		gen1 := &site.Site{Content: hash.Hash{24, 67, 68, 72, 132, 181}, Nonce: 373}
		gen2 := &site.Site{Content: hash.Hash{24, 67, 68, 72, 132, 182}, Nonce: 510}
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
	} else {
		log.Info("Restoring Tangle")
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
func (t *Tangle) Add(s *site.Site) error {
	err := t.verifySite(s)
	if err != nil {
		return err
	}
	v := func() bool {
		for _, v := range s.Validates {
			if t.hasTip(v) {
				return true
			}
		}
		return false
	}()
	if !v {
		return ErrNotValidating
	}
	for _, vs := range s.Validates {
		delete(t.tips, vs)
	}
	t.tips[s] = true
	t.store.SetTips(s, s.Validates)
	t.addSite(s)
	return nil
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
func (t *Tangle) Get(h hash.Hash) *site.Site {
	return t.store.Get(h)
}

// Close closes the underlying store
func (t *Tangle) Close() {
	t.store.Close()
}

func (t *Tangle) hasTip(s *site.Site) bool {
	return t.tips[s]
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

func (t *Tangle) addSite(s *site.Site) {
	err := t.store.Add(s)
	if err != nil {
		log.Error(err)
	}
}
