package node

import (
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
)

type Node struct {
	Chain           *chain.Chain
	ListenInterface string
}

type Status struct {
	Address     string `json:"address"`
	Version     string `json:"version"`
	Length      int    `json:"length"`
	Connections int    `json:"connections"`
}

func validateAll([32]byte) bool {
	return true
}

func New(c config.Configuration) *Node {
	return &Node{
		ListenInterface: c.NodeNetwork.Interface + ":" + strconv.Itoa(c.NodeNetwork.Port),
		Chain:           chain.New(&chain.MemoryStore{}, validateAll),
	}
}

func (n *Node) Status() Status {
	return Status{
		Address: n.ListenInterface,
	}
}

func (n *Node) Run() {
	log.Debug("Simulating a running server")
	for {
	}
}

func (n *Node) SubmitBlock(b chain.Block) {
	log.Infof("Pushing block %x to network", b.Hash())
}
