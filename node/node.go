package node

import (
	"container/list"
	"encoding/base64"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
	d "github.com/u-speak/core/node/protoc"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
	"time"
)

// Node is a wrapper around the chain. Nodes are the backbone of the network
type Node struct {
	PostChain         *chain.Chain
	ImageChain        *chain.Chain
	KeyChain          *chain.Chain
	ListenInterface   string
	Version           string
	remoteConnections map[string]*grpc.ClientConn
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
	Connections []string        `json:"connections"`
	Chains      ChainStatusList `json:"chains"`
}

func validateAll([32]byte) bool {
	return true
}

// New constructs a new node from the configuration
func New(c config.Configuration) (*Node, error) {
	ic, err := chain.New(&chain.BoltStore{Path: c.Storage.BoltStore.ImagePath}, validateAll)
	if err != nil {
		return nil, err
	}
	kc, err := chain.New(&chain.BoltStore{Path: c.Storage.BoltStore.KeyPath}, validateAll)
	if err != nil {
		return nil, err
	}
	pc, err := chain.New(&chain.BoltStore{Path: c.Storage.BoltStore.PostPath}, validateAll)
	if err != nil {
		return nil, err
	}
	return &Node{
		ListenInterface:   c.NodeNetwork.Interface + ":" + strconv.Itoa(c.NodeNetwork.Port),
		ImageChain:        ic,
		KeyChain:          kc,
		PostChain:         pc,
		Version:           c.Version,
		remoteConnections: make(map[string]*grpc.ClientConn),
	}, nil
}

func encHash(h [32]byte) string {
	return base64.URLEncoding.EncodeToString(h[:])
}

// Status returns the current running configuration of the node
func (n *Node) Status() Status {
	cons := []string{}
	for k := range n.remoteConnections {
		cons = append(cons, k)
	}
	return Status{
		Address: n.ListenInterface,
		Length:  n.PostChain.Length() + n.KeyChain.Length() + n.ImageChain.Length(),
		Chains: ChainStatusList{
			Post:  ChainStatus{Length: n.PostChain.Length(), Valid: n.PostChain.Valid(), LastHash: encHash(n.PostChain.LastHash())},
			Image: ChainStatus{Length: n.ImageChain.Length(), Valid: n.ImageChain.Valid(), LastHash: encHash(n.ImageChain.LastHash())},
			Key:   ChainStatus{Length: n.KeyChain.Length(), Valid: n.KeyChain.Valid(), LastHash: encHash(n.KeyChain.LastHash())},
		},
		Connections: cons,
		Version:     n.Version,
	}
}

// GetInfo is a all purpose status request
func (n *Node) GetInfo(ctx context.Context, params *d.StatusParams) (*d.Info, error) {
	if _, contained := n.remoteConnections[params.Host]; !contained {
		err := n.Connect(params.Host)
		if err != nil {
			log.Error("Failed to initialize reverse connection. Fulfilling request anyways...")
		}
	}
	return &d.Info{
		Length: n.PostChain.Length(),
	}, nil
}

// Run listens for connections to this node
func (n *Node) Run() {
	log.Infof("Starting Nodeserver on %s", n.ListenInterface)
	lis, err := net.Listen("tcp", n.ListenInterface)
	if err != nil {
		log.Errorf("Could not listen on %s: %s", n.ListenInterface, err)
	}
	grpcServer := grpc.NewServer()
	d.RegisterDistributionServiceServer(grpcServer, n)
	log.Fatal(grpcServer.Serve(lis))
}

// Connect connects to a new remote
func (n *Node) Connect(remote string) error {
	if _, contained := n.remoteConnections[remote]; contained {
		return errors.New("Node allready connected")
	}
	conn, err := grpc.Dial(remote, grpc.WithInsecure())
	if err != nil {
		return err
	}
	n.remoteConnections[remote] = conn
	log.Infof("Successfully connected to %s", remote)
	return nil
}

// SubmitBlock is called whenever a new block is submitted to the network
func (n *Node) SubmitBlock(b chain.Block) {
	log.Debug(n.PostChain)
	log.Infof("Pushing block %x to network", b.Hash())
	n.Push(&b)
}

// Push sends a block to all connected nodes
func (n *Node) Push(b *chain.Block) {
	h := b.PrevHash
	pb := &d.Block{
		Content:   b.Content,
		Nonce:     b.Nonce,
		Previous:  h[:],
		Signature: b.Signature,
		Date:      b.Date.Unix(),
		Type:      b.Type,
		PubKey:    b.PubKey,
	}
	for _, r := range n.remoteConnections {
		client := d.NewDistributionServiceClient(r)
		_, err := client.AddBlock(context.Background(), pb)
		if err != nil {
			log.Error(err)
		}
	}
}

// AddBlock receives a sent Block from other node
func (n *Node) AddBlock(ctx context.Context, block *d.Block) (*d.PushReturn, error) {
	log.Debugf("Received Block: %s", block.Content)
	var p [32]byte
	copy(p[:], block.Previous)
	if p != n.PostChain.LastHash() {
		log.Errorf("Tried to add invalid Block! Previous hash %v is not valid. Please synchronize the nodes", p)
		return &d.PushReturn{}, errors.New("Received block had invalid previous hash")
	}
	var h [32]byte
	copy(h[:], block.Previous)
	b := chain.Block{
		Content:   block.Content,
		Type:      block.Type,
		PubKey:    block.PubKey,
		Date:      time.Unix(int64(block.Date), 0),
		Signature: block.Signature,
		PrevHash:  h,
		Nonce:     block.Nonce,
	}
	var c *chain.Chain
	switch b.Type {
	case "post":
		c = n.PostChain
	case "image":
		c = n.ImageChain
	case "key":
		c = n.KeyChain
	}
	c.Add(b)
	return &d.PushReturn{}, nil
}

// Synchronize sends all Blocks to an other node
func (n *Node) Synchronize(p *d.SyncParams, stream d.DistributionService_SynchronizeServer) error {
	log.Infof("Synchronization started. Sending all Blocks to another Node.")
	h := n.PostChain.LastHash()
	b := n.PostChain.Get(h)

	c := [32]byte{}
	var blst list.List
	for {
		blst.PushBack(b.Content)
		if b.PrevHash == c {
			break
		}
		b = n.PostChain.Get(b.PrevHash)
	}
	blk := []*chain.Block{}
	blk, _ = n.PostChain.DumpChain()

	for i := len(blk) - 2; i >= 0; i-- {
		err := stream.Send(&d.Block{
			Content:   blk[i].Content,
			Nonce:     blk[i].Nonce,
			Previous:  blk[i].PrevHash[:],
			Type:      blk[i].Type,
			PubKey:    blk[i].PubKey,
			Date:      blk[i].Date.Unix(),
			Signature: blk[i].Signature,
		})
		if err != nil {
			log.Error(err)
		}
		log.Debugf("Sent %+v", blk[i])

	}
	log.Infof("Synchronization finished successfully.")
	return nil
}

// SynchronizeChain receives all the Blocks sent from an other node
func (n *Node) SynchronizeChain(remote string) error {
	lh := n.PostChain.LastHash()
	log.Infof("Synchronization started. Receiving Blocks from other node.")
	params := &d.SyncParams{LastHash: lh[:]}
	client := d.NewDistributionServiceClient(n.remoteConnections[remote])
	stream, err := client.Synchronize(context.Background(), params)
	if err != nil {
		return err
	}
	for {
		block, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		var p [32]byte
		copy(p[:], block.Previous)

		b := chain.Block{
			Content:   block.Content,
			Type:      "post",
			PubKey:    block.PubKey,
			Date:      time.Unix(block.Date, 0),
			Signature: block.Signature,
			PrevHash:  p,
			Nonce:     block.Nonce,
		}
		log.Infof("Got a new Block: %v", b.Content)
		_, err = n.PostChain.Add(b)
		if err != nil {
			return err
		} else {
			log.Debug("added %s", b.Content)
		}
	}
	log.Infof("Synchronization finished successfully.")
	return nil
}
