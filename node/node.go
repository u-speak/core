package node

import (
	"encoding/base64"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
	d "github.com/u-speak/core/node/protoc"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"strconv"
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
	return Status{
		Address: n.ListenInterface,
		Length:  n.PostChain.Length() + n.KeyChain.Length() + n.ImageChain.Length(),
		Chains: ChainStatusList{
			Post:  ChainStatus{Length: n.PostChain.Length(), Valid: n.PostChain.Valid(), LastHash: encHash(n.PostChain.LastHash())},
			Image: ChainStatus{Length: n.ImageChain.Length(), Valid: n.ImageChain.Valid(), LastHash: encHash(n.ImageChain.LastHash())},
			Key:   ChainStatus{Length: n.KeyChain.Length(), Valid: n.KeyChain.Valid(), LastHash: encHash(n.KeyChain.LastHash())},
		},
		Connections: len(n.remoteConnections),
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
		Nonce:     uint32(b.Nonce),
		Previous:  h[:],
		Signature: b.Signature,
		Date:      uint32(b.Date.Unix()),
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

// Receive receives a sent Block from other node, also just PostChain Blocks at the moment
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
		Nonce:     uint(block.Nonce),
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

func (n *Node) Synchronize(p *d.SyncParams, stream d.DistributionService_SynchronizeServer) error {
	h := n.PostChain.LastHash()
	b := n.PostChain.Get(h)


	var c [32]byte
	copy(c[:], p.LastHash)
	for {
		if err := stream.Send(&d.Block{Content: b.Content, Type: b.Type, PubKey: b.PubKey, Date: uint32(b.Date.Unix()), Signature: b.Signature, Previous:  b.PrevHash[:]}); err != nil {
			log.Error(err)
		}
		if b.PrevHash == c {
			break
		}
		b = n.PostChain.Get(b.PrevHash)
	}
	return nil
}
