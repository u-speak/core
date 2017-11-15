package chain

// BlockStore is the interface needed for storing data
type BlockStore interface {
	Init() ([32]byte, error)
	Get([32]byte) *Block
	Add(Block) error
	Length() uint64
	//Keys() [][32]byte
	Valid(func([32]byte) bool) bool
	Initialized() bool
	Close()
}
