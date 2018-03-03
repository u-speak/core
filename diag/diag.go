package diag

//go:generate go-bindata -pkg diag -nomemcopy static/...
import (
	"net/http"
	"strconv"

	"github.com/u-speak/core/config"
	"github.com/u-speak/core/node"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/logrusmiddleware"
)

// Server stores the state for the diagnostics server
type Server struct {
	node *node.Node
}

type edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Run starts the diagnostics server
func Run(c config.Configuration, n *node.Node) error {
	s := Server{node: n}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = logrusmiddleware.Logger{log.StandardLogger()}
	e.GET("/", s.getIndex)
	e.GET("/static/:name", s.getStatic)
	e.GET("/tangle/graph", s.getGraph)
	return e.StartTLS(c.Diagnostics.Interface+":"+strconv.Itoa(c.Diagnostics.Port), c.Global.SSLCert, c.Global.SSLKey)
}

func (s *Server) getIndex(c echo.Context) error {
	b, err := Asset("static/index.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.HTMLBlob(http.StatusOK, b)
}

func (s *Server) getGraph(c echo.Context) error {
	hs := s.node.Tangle.Hashes()
	bound := make(map[*site.Site]bool)
	ti := s.node.Tangle.Tips()
	for _, t := range ti {
		bound[t] = true
	}
	excl := make(map[hash.Hash]bool)
	edges := []edge{}
	for len(bound) > 0 {
		for s := range bound {
			sh := s.Hash()
			excl[sh] = true
			delete(bound, s)
			for _, v := range s.Validates {
				edges = append(edges, edge{From: sh.String(), To: v.Hash().String()})
				if !excl[v.Hash()] {
					bound[v] = true
				}
			}
		}
	}
	hss := []string{}
	for _, h := range hs {
		hss = append(hss, h.String())
	}
	return c.JSON(http.StatusOK, struct {
		Nodes []string `json:"nodes"`
		Edges []edge   `json:"edges"`
	}{Nodes: hss, Edges: edges})
}

func (s *Server) getStatic(c echo.Context) error {
	b, err := Asset("static/" + c.Param("name"))
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	return c.Blob(http.StatusOK, "", b)
}
