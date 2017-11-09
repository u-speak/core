package node

import (
	"encoding/base64"
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

type ChainStatus struct {
	Valid    bool   `json:"valid"`
	Length   uint64 `json:"length"`
	LastHash string `json:"last_hash"`
}

type ChainStatusList struct {
	Post  ChainStatus `json:"post"`
	Image ChainStatus `json:"image"`
	Key   ChainStatus `json:"key"`
}

// Status is used for reporting this nodes configuration to other nodes
type Status struct {
	Address     string          `json:"address"`
	Version     string          `json:"version"`
	Length      uint64          `json:"length"`
	Connections int             `json:"connections"`
	Chains      ChainStatusList `json:"chains"`
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

func encHash(h [32]byte) string {
	return base64.URLEncoding.EncodeToString(h[:])
}

// Status returns the current running configuration of the node
func (n *Node) Status() Status {
	return Status{
		Address: n.ListenInterface,
		Length:  n.PostChain.Length() + n.KeyChain.Length() + n.ImageChain.Length(),
		Chains: ChainStatusList{
			Post:  ChainStatus{Length: n.PostChain.Length(), Valid: n.PostChain.Valid(), LastHash: encHash(n.PostChain.LastHash())},
			Image: ChainStatus{Length: n.ImageChain.Length(), Valid: n.ImageChain.Valid(), LastHash: encHash(n.ImageChain.LastHash())},
			Key:   ChainStatus{Length: n.KeyChain.Length(), Valid: n.KeyChain.Valid(), LastHash: encHash(n.KeyChain.LastHash())},
		},
	}
}

// Run listens for connections to this nodgit ammende
func (n *Node) Run() {
	log.Debug("Simulating a running server")
}

// SubmitBlock is called whenever a new block is submitted to the network
func (n *Node) SubmitBlock(b chain.Block) {
	log.Infof("Pushing block %x to network", b.Hash())
}
