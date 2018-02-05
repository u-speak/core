package tangle

import (
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/store"
	"github.com/u-speak/core/tangle/store/boltstore"
)

const (
	// MinimumWeight for new Sites
	MinimumWeight = 1
	// MinimumValidations specifies how many sites must be verified by a new site
	MinimumValidations = 2
)

// Tangle stores the relation between different transactions
type Tangle struct {
	tips  map[*Site]bool
	sites map[hash.Hash]*Site
	store store.Store
}

// Init initializes the tangle with two genesis blocks
func (t *Tangle) Init() error {
	t.tips = make(map[*Site]bool)
	t.sites = make(map[hash.Hash]*Site)
	// Base64(Content) = GENESIS1
	g1 := &Site{Content: hash.Hash{24, 67, 68, 72, 132, 181}, Nonce: 373}
	// Base64(Content) = GENESIS2
	g2 := &Site{Content: hash.Hash{24, 67, 68, 72, 132, 182}, Nonce: 510}
	t.sites[g1.Hash()] = g1
	t.sites[g2.Hash()] = g2
	t.tips[g1] = true
	t.tips[g2] = true
	return nil
}

// Add Validates the site and adds it to the tangle
// to be valid, a site has to:
// * Validate at least one tip
// * Have a weight of at least MinimumWeight
func (t *Tangle) Add(s *Site) error {
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
	t.addSite(s)
	return nil
}

// Size returns the amount of sites in the tangle
func (t *Tangle) Size() int {
	return len(t.sites)
}

// Tips returns a list of unconfirmed tips
func (t *Tangle) Tips() []*Site {
	keys := []*Site{}
	for s := range t.tips {
		keys = append(keys, s)
	}
	return keys
}

// Get retrieves the specified site
func (t *Tangle) Get(h hash.Hash) *Site {
	return t.sites[h]
}

func (t *Tangle) hasTip(s *Site) bool {
	return t.tips[s]
}

func (t *Tangle) verifySite(s *Site) error {
	if s.Hash().Weight() < MinimumWeight {
		return ErrWeightTooLow
	}
	if len(s.Validates) < MinimumValidations {
		return ErrTooFewValidations
	}
	return nil
}

func (t *Tangle) addSite(s *Site) {
	t.sites[s.Hash()] = s
}
