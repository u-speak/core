package chain

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// DiskStore is a Blockstore implementation saving the blocks serialized to a Folder
type DiskStore struct {
	Folder      string
	initialized bool
}

// Init initializes the Diskstore
func (b *DiskStore) Init() ([32]byte, error) {
	err := os.Mkdir(b.Folder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return [32]byte{}, err
	}
	ph := map[[32]byte]bool{}
	a := b.all()
	if len(a) == 0 {
		log.Infof("Initializing empty chain in directory %s", b.Folder)
		g := genesisBlock()
		err := b.Add(g)
		if err != nil {
			return [32]byte{}, err
		}
		return g.Hash(), nil
	}
	for _, bl := range a {
		ph[bl.PrevHash] = true
	}
	for _, bl := range a {
		if ph[bl.Hash()] == false {
			return bl.Hash(), nil
		}
	}
	b.initialized = true
	return [32]byte{}, errors.New("Could not calculate lasthash")
}

// Initialized returns whether or not this store has been initialized
func (b *DiskStore) Initialized() bool {
	return b.initialized
}

// Get retrieves a block by its hash
func (b *DiskStore) Get(hash [32]byte) *Block {
	file, err := os.Open(path.Join(b.Folder, base64.URLEncoding.EncodeToString(hash[:])))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		log.Error(err)
		return nil
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	bl := &Block{}
	err = dec.Decode(bl)
	if err != nil {
		log.Error(err)
		return nil
	}
	if bl.Hash() != hash {
		log.Error("Tried to load a modified block")
		return nil
	}
	return bl
}

// Add adds a block to the raw storage
func (b *DiskStore) Add(block Block) error {
	h := block.Hash()
	file, err := os.Create(path.Join(b.Folder, base64.URLEncoding.EncodeToString(h[:])))
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(block)
	if err != nil {
		return err
	}
	return nil
}

// Length returns the length of the whole chain
func (b *DiskStore) Length() uint64 {
	return uint64(len(b.all()))
}

func (b *DiskStore) all() []*Block {
	all := []*Block{}
	files, err := ioutil.ReadDir(b.Folder)
	if err != nil {
		log.Error(err)
		return nil
	}
	for _, f := range files {
		stat := [32]byte{}
		h, err := base64.URLEncoding.DecodeString(f.Name())
		if err != nil {
			log.Warn(err)
			continue
		}
		copy(stat[:], h)
		bl := b.Get(stat)
		if bl != nil {
			all = append(all, bl)
		}
	}
	return all
}

// Valid checks if all blocks are connected and have the required difficulty
func (b *DiskStore) Valid(v func([32]byte) bool) bool {
	a := b.all()
	if len(a) == 0 {
		return false
	}
	f := make(map[[32]byte]bool)
	for _, h := range a {
		f[h.Hash()] = true
	}
	for _, b := range a {
		if !v(b.Hash()) {
			return false
		} else if !f[b.PrevHash] && b.Content != "GENESIS" {
			return false
		}
	}
	return true
}

// Close closes the underlying connections
func (b *DiskStore) Close() {
}
