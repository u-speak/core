package node

import (
<<<<<<< HEAD
	"errors"
=======
	"encoding/base64"
>>>>>>> a3ecea7a5799f548f417464b5a408a2b3a4cc022
	"strconv"
	"google.golang.org/grpc"
	"net"
	d "github.com/u-speak/core/node/protoc"
	context "golang.org/x/net/context"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
)

//Node is a wrapper around the chain. Nodes are the backbone of the network
type Node struct {
	PostChain       *chain.Chain
	ImageChain      *chain.Chain
	KeyChain        *chain.Chain
	ListenInterface string
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
//	fmt.Println(config.NodeNetwork.Interface)
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

<<<<<<< HEAD

func (s *Node) GetInfo(ctx context.Context, params *d.StatusParams) (*d.Info, error) {
	if _, contained := s.remoteConnections[params.Host]; !contained {
		err := s.Connect(params.Host)
		if err != nil {
			log.Error("Failed to initialize reverse connection. Fulfilling request anyways...")
		}
	}
	//lh := s.PostChain.LastHash()
	return &d.Info{
		Length:   s.PostChain.Length(),
			}, nil
}


// Run listens for connections to this node
=======
// Run listens for connections to this nodgit ammende
>>>>>>> a3ecea7a5799f548f417464b5a408a2b3a4cc022
func (n *Node) Run() {
//	fmt.Println(config.NodeNetwork.Interface)
//	fmt.Println(config.NodeNetwork.Interface)
	log.Infof("Starting Nodeserver on 127.0.0.1:6969")
	lis, _ := net.Listen("tcp", "127.0.0.1:6969")
        grpcServer := grpc.NewServer()
        d.RegisterDistributionServiceServer(grpcServer, n)
        log.Fatal(grpcServer.Serve(lis))

        log.Infof("Started Nodeserver.")
}

func (s *Node) Connect(remote string) error {
        if _, contained := s.remoteConnections[remote]; contained {
                return errors.New("Node allready connected")
        }
        conn, err := grpc.Dial(remote, grpc.WithInsecure())
        if err != nil {
                return err
        }
        s.remoteConnections[remote] = conn
        log.Infof("Successfully connected to %s", remote)
        return nil
}


// SubmitBlock is called whenever a new block is submitted to the network
func (n *Node) SubmitBlock(b chain.Block) {
	log.Debug(n.PostChain)
	log.Infof("Pushing block %x to network", b.Hash())
	n.PostChain.Add(b)
}

//Send a block to all node which are currently connected, just PostChain Blocks at the moment
func (s *Node) Push(b *chain.Block) {
	lh := s.PostChain.LastHash()
	for _, r := range s.remoteConnections {
		client := d.NewDistributionServiceClient(r)
		_, err := client.Receive(context.Background(), &d.Block{Content: b.Content, Nonce: uint32(b.Nonce), Previous: lh[:]})
		if err != nil {
			log.Error(err)
		}
	}
}

//Receives a sent Block from other node, also just PostChain Blocks at the moment
func (s *Node) Receive(ctx context.Context, block *d.Block) (*d.PushReturn, error) {
        log.Debugf("Received Block: %s", block.Content)
        var p [32]byte
        copy(p[:], block.Previous)
        if p != s.PostChain.LastHash() {
                log.Errorf("Tried to add invalid Block! Previous hash %v is not valid. Please synchronize the nodes", p)
                return &d.PushReturn{}, errors.New("Received block had invalid previous hash")
        }
        return &d.PushReturn{}, nil
}


func (s *Node) Synchronize(p *d.SyncParams, stream d.DistributionService_SynchronizeServer) error {
	h := s.PostChain.LastHash()
	b := s.PostChain.Get(h)
	var c [32]byte
	copy(c[:], p.LastHash)
	last := s.PostChain.LastHash()
	for {
		if err := stream.Send(&d.Block{Content: b.Content, Nonce: uint32(b.Nonce), Previous: b.PrevHash[:], Last: last[:]}); err != nil {
			log.Error(err)
		}
		if b.PrevHash == c {
			break
		}
		b = s.PostChain.Get(b.PrevHash)
	}
	return nil
}
