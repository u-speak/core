package chain

import (
	"testing"
)

func TestReinitialize(t *testing.T) {
	cases := []BlockStore{
		emptyDiskStore(),
		&MemoryStore{},
		emptyBoltStore(),
	}
	for _, c := range cases {
		ih, err := c.Init()
		if err != nil {
			t.Error("Initialization failed: ", err)
		}
		err = c.Add(Block{Content: "foo", PrevHash: ih})
		if err != nil {
			t.Error("Error adding block: ", err)
		}
		nh, err := c.Reinitialize()
		if err != nil {
			t.Error("Reinitialization failed: ", err)
		}
		if nh != ih {
			t.Logf("%+v %+v", nh, ih)
			t.Errorf("Reinitialized store did not share the same genesis block. Store: %+v", c)
		}

	}
}
