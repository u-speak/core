package node

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/u-speak/core/config"
	"github.com/u-speak/core/img"
	"github.com/u-speak/core/post"
	"github.com/u-speak/core/tangle"
	"github.com/u-speak/core/tangle/datastore"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"
	"github.com/u-speak/core/tangle/store"
	"github.com/u-speak/core/tangle/store/boltstore"

	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
	d "github.com/u-speak/core/node/internal"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// MaxMsgSize specifies the largest packet size for grpc calls
	MaxMsgSize = 5242880
)

// Node is a wrapper around the chain. Nodes are the backbone of the network
type Node struct {
	Tangle           *tangle.Tangle
	ListenInterface  string
	Version          string
	remoteInterfaces map[string]struct{}
	APIAddr          string
	Hooks            struct {
		PreAdd string
	}
}

// Status is used for reporting this nodes configuration to other nodes
type Status struct {
	Address     string      `json:"address"`
	Version     string      `json:"version"`
	Length      uint64      `json:"length"`
	Connections []string    `json:"connections"`
	Hashes      []hash.Hash `json:"hashes"`
	HashDiff    HashDiff
}

// HashDiff stores the diff between two tangles
type HashDiff struct {
	Additions []hash.Hash
	Deletions []hash.Hash
}

// New constructs a new node from the configuration
func New(c config.Configuration) (*Node, error) {
	n := &Node{
		ListenInterface:  c.NodeNetwork.Interface + ":" + strconv.Itoa(c.NodeNetwork.Port),
		Version:          c.Version,
		remoteInterfaces: make(map[string]struct{}),
		Hooks:            c.Hooks,
		APIAddr:          c.Web.API.PublicEndpoint,
	}
	bs, err := boltstore.New(store.Options{Path: c.Storage.TanglePath})
	if err != nil {
		return nil, err
	}
	tngl, err := tangle.New(tangle.Options{Store: bs, DataPath: c.Storage.DataPath})
	n.Tangle = tngl
	return n, err
}

// Status returns the current running configuration of the node
func (n *Node) Status() Status {
	cons := []string{}
	for k := range n.remoteInterfaces {
		cons = append(cons, k)
	}
	return Status{
		Address:     n.ListenInterface,
		Length:      uint64(n.Tangle.Size()),
		Connections: cons,
		Version:     n.Version,
		Hashes:      n.Tangle.Hashes(),
	}
}

// RemoteStatus returns the status of a connected remote
func (n *Node) RemoteStatus(s string) (*Status, error) {
	conn, err := dial(s)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := d.NewDistributionServiceClient(conn)
	i, err := client.GetInfo(context.Background(), n.Info())
	if err != nil {
		return nil, err
	}
	hs := []hash.Hash{}
	for _, h := range i.Hashes {
		hs = append(hs, hash.FromSlice(h))
	}
	a, d := hash.Diff(n.Tangle.Hashes(), hs)
	return &Status{
		Version:     i.Version,
		Length:      i.Length,
		Connections: i.Connections,
		Address:     i.ListenInterface,
		Hashes:      hs,
		HashDiff:    HashDiff{Additions: a, Deletions: d},
	}, nil
}

// Info returns the serializable info struct
func (n *Node) Info() *d.Info {
	s := n.Status()
	cons := []string{}
	for k := range n.remoteInterfaces {
		cons = append(cons, k)
	}
	hs := [][]byte{}
	for _, h := range n.Tangle.Hashes() {
		hs = append(hs, h.Slice())
	}
	return &d.Info{
		Length:          s.Length,
		ListenInterface: s.Address,
		Version:         n.Version,
		Connections:     cons,
		Hashes:          hs,
	}
}

// GetInfo is a all purpose status request
func (n *Node) GetInfo(ctx context.Context, r *d.Info) (*d.Info, error) {
	if _, ok := n.remoteInterfaces[r.ListenInterface]; !ok && n.ListenInterface != r.ListenInterface {
		log.Infof("Establishing reverse connection with %s", r.ListenInterface)
		n.Connect(r.ListenInterface)
	}
	return n.Info(), nil
}

// Run listens for connections to this node
func (n *Node) Run() {
	log.Infof("Starting Nodeserver on %s", n.ListenInterface)
	lis, err := net.Listen("tcp", n.ListenInterface)
	if err != nil {
		log.Errorf("Could not listen on %s: %s", n.ListenInterface, err)
	}
	// Set MsgSize to 5MB
	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(MaxMsgSize), grpc.MaxRecvMsgSize(MaxMsgSize))
	d.RegisterDistributionServiceServer(grpcServer, n)

	log.Info("Starting cronjobs")
	go n.startCron()
	log.Fatal(grpcServer.Serve(lis))
}

func (n *Node) startCron() {
	gocron.Every(1).Minute().Do(func() {
		for r := range n.remoteInterfaces {
			s, err := n.RemoteStatus(r)
			if err != nil {
				log.Error(err)
				continue
			}
			if len(s.HashDiff.Additions) == 0 && len(s.HashDiff.Deletions) == 0 {
				continue
			}
			err = n.Merge(r)
			if err != nil {
				log.Error(err)
			}
		}
	})
	<-gocron.Start()
}

func (n *Node) connect(remote string) error {
	if _, ok := n.remoteInterfaces[remote]; ok {
		return errors.New("Attempted to add an allready established interface")
	}
	n.remoteInterfaces[remote] = struct{}{}
	conn, err := dial(remote)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := d.NewDistributionServiceClient(conn)
	_, err = client.GetInfo(context.Background(), n.Info())
	if err != nil {
		delete(n.remoteInterfaces, remote)
		return err
	}
	n.remoteInterfaces[remote] = struct{}{}
	log.Infof("Added connection %s", remote)
	return nil
}

// Connect connects to a new remote
func (n *Node) Connect(r string) error {
	s := strings.Split(r, ":")
	port := s[1]
	addr := s[0]
	i, err := net.LookupIP(addr)
	if err != nil {
		return err
	}
	for _, ip := range i {
		if ip.To4() != nil {
			err := n.connect(ip.String() + ":" + port)
			if err != nil {
				log.Error(err)
			}
		} else {
			log.Warn("Not using IPv6 as of now")
		}
	}
	return nil
}

// Submit is called whenever a new site is submitted to the network
func (n *Node) Submit(o *tangle.Object) error {
	log.Infof("Pushing site %s to network", o.Site.Hash())
	return n.Push(o)
}

// Push sends a site to all connected nodes
func (n *Node) Push(o *tangle.Object) error {
	ds, err := d.FromObject(o)
	if err != nil {
		return err
	}
	for r := range n.remoteInterfaces {
		conn, err := dial(r)
		if err != nil {
			log.Error(err)
			continue
		}
		defer conn.Close()
		client := d.NewDistributionServiceClient(conn)
		_, err = client.AddSite(context.Background(), ds)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// AddSite receives a sent Site from other node
func (n *Node) AddSite(ctx context.Context, s *d.Site) (*d.SuccessReturn, error) {
	o, err := n.toObject(s)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debugf("Received Site %s", o.Site.Hash())
	if n.Hooks.PreAdd != "" {
		u, err := url.Parse(n.Hooks.PreAdd)
		if err != nil {
			log.Errorf("Error running PreAdd hook: %s", err.Error())
		}
		q := u.Query()
		q.Add("hash", base64.URLEncoding.EncodeToString(o.Site.Hash().Slice()))
		q.Add("pub", n.APIAddr)
		u.RawQuery = q.Encode()
		log.Debugf("Calling PreAdd Hook with URL: %s", u.String())
		_, err = http.Get(u.String())
		if err != nil {
			log.Errorf("Error running PreAdd hook: %s", err.Error())
		}
	}
	err = n.Tangle.Inject(o, true)
	if err != nil {
		log.Errorf("Failed to add site: %s", err)
	} else {
		log.Infof("Successfully added site: %s", o.Site.Hash())
	}
	return &d.SuccessReturn{}, err
}

// Merge requests to merge with a remote
func (n *Node) Merge(r string) error {
	s, err := n.RemoteStatus(r)
	if err != nil {
		return err
	}
	if len(s.HashDiff.Additions) == 0 && len(s.HashDiff.Deletions) == 0 {
		return errors.New("Nodes are up to date - No merge needed")
	}
	log.Infof("Merge Summary: %d local additions, %d remote additions", len(s.HashDiff.Additions), len(s.HashDiff.Deletions))
	conn, err := dial(r)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := d.NewDistributionServiceClient(conn)
	stream, err := client.Splice(context.Background())
	if err != nil {
		return err
	}
	for _, h := range s.HashDiff.Deletions {
		o := n.Tangle.Get(h)
		if o == nil {
			continue
		}
		do, err := d.FromObject(o)
		if err != nil {
			return err
		}
		if n.Tangle.HasTip(o.Site.Hash()) {
			do.Tip = true
		}
		err = stream.Send(do)
		if err != nil {
			return err
		}
		log.Infof("Sent %s", o.Site.Hash())
	}
	_, err = stream.CloseAndRecv()
	if err == io.EOF {
		return nil
	}
	return err
}

// Splice injects the recieved sites into the tangle
func (n *Node) Splice(stream d.DistributionService_SpliceServer) error {
	canLink := func(o *d.Site) bool {
		for _, s := range o.Validates {
			h := hash.FromSlice(s)
			if n.Tangle.Get(h) == nil {
				return false
			}
		}
		return true
	}
	inj := func(o *d.Site) error {
		s, err := n.toObject(o)
		if err != nil {
			return err
		}
		log.Infof("Received Site %s", s.Site.Hash())
		err = n.Tangle.Inject(s, o.Tip)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	log.Info("Starting Splice")
	buff := make(map[*d.Site]bool)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Info("Finished Splicing")
			break
		}
		if err != nil {
			log.Error(err)
			return err
		}
		if canLink(in) {
			err := inj(in)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			buff[in] = true
		}
	}
	log.Infof("Remaining injections: %d", len(buff))
	for len(buff) > 0 {
		origlen := len(buff)
		for s := range buff {
			if canLink(s) {
				err := inj(s)
				if err != nil {
					log.Error(err)
					return err
				}
				delete(buff, s)
			}
		}
		if len(buff) == origlen {
			return errors.New("Merge Failed! Invalid tangle structure")
		}
	}
	return nil
}

func (n *Node) toObject(s *d.Site) (*tangle.Object, error) {
	vs := []*site.Site{}
	for _, h := range s.Validates {
		o := n.Tangle.Get(hash.FromSlice(h))
		if o == nil {
			return nil, errors.New("This node does not know about hash " + hash.FromSlice(h).String())
		}
		vs = append(vs, o.Site)
	}
	var d datastore.Serializable
	switch s.Type {
	case "post":
		d = &post.Post{}
	case "image":
		d = &img.Image{}
	default:
		return nil, errors.New("Invalid site type")
	}
	err := d.Deserialize(s.Data)
	if err != nil {
		return nil, err
	}
	return &tangle.Object{
		Site: &site.Site{
			Validates: vs,
			Nonce:     s.Nonce,
			Content:   hash.FromSlice(s.Content),
			Type:      s.Type,
		},
		Data: d,
	}, nil
}

func dial(r string) (*grpc.ClientConn, error) {
	return grpc.Dial(r,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxMsgSize),
			grpc.MaxCallSendMsgSize(MaxMsgSize),
		))
}
