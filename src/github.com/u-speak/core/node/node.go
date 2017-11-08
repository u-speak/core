package node

import (
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
)

// Node is a wrapper around the chain. Nodes are the backbone of the network
type Node struct {
	PostChain       *chain.Chain
	ImageChain      *chain.Chain
	KeyChain        *chain.Chain
	ListenInterface string
}

// Status is used for reporting this nodes configuration to other nodes
type Status struct {
	Address     string `json:"address"`
	Version     string `json:"version"`
	Length      uint64 `json:"length"`
	Connections int    `json:"connections"`
}

func validateAll([32]byte) bool {
	return true
}

// New constructs a new node from the configuration
func New(c config.Configuration) (*Node, error) {
	ic, err := chain.New(&chain.DiskStore{Folder: c.Storage.ImageDir}, validateAll)
	if err != nil {
		return nil, err
	}
	kc, err := chain.New(&chain.DiskStore{Folder: c.Storage.KeyDir}, validateAll)
	if err != nil {
		return nil, err
	}
	pc, err := chain.New(&chain.DiskStore{Folder: c.Storage.PostDir}, validateAll)
	if err != nil {
		return nil, err
	}
	return &Node{
		ListenInterface: c.NodeNetwork.Interface + ":" + strconv.Itoa(c.NodeNetwork.Port),
		ImageChain:      ic,
		KeyChain:        kc,
		PostChain:       pc,
	}, nil
}

// Status returns the current running configuration of the node
func (n *Node) Status() Status {
	return Status{
		Address: n.ListenInterface,
		Length:  n.PostChain.Length() + n.KeyChain.Length() + n.ImageChain.Length(),
	}
}

// Run listens for connections to this node
func (n *Node) Run() {
	log.Debug("Simulating a running server")
}

// SubmitBlock is called whenever a new block is submitted to the network
func (n *Node) SubmitBlock(b chain.Block) {
	log.Infof("Pushing block %x to network", b.Hash())
}
