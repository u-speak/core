package chain

// BlockStore is the interface needed for storing data
type BlockStore interface {
	Init() (Hash, error)
	Get(Hash) *Block
	Add(Block) error
	Length() uint64
	//Keys() [][32]byte
	Valid(func(Hash) bool) bool
	Initialized() bool
	Close()
	Reinitialize() (Hash, error)
}
