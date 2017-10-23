package api

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/kpashka/echo-logrusmiddleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/node"
)

// API is used as a container, allowing the REST API to access the node
type API struct {
	ListenInterface string
	node            *node.Node
	certfile        string
	keyfile         string
}

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type jsonBlock struct {
	Nonce     uint   `json:"nonce"`
	PrevHash  string `json:"previous_hash"`
	Hash      string `json:"hash"`
	Content   string `json:"content"`
	Signature string `json:"signature"`
	Type      string `json:"type"`
	PubKey    string `json:"public_key"`
	Date      int64  `json:"date"`
}

// New returns a configured instance of the API server
func New(c config.Configuration, n *node.Node) *API {
	a := &API{node: n, keyfile: c.Global.SSLKey, certfile: c.Global.SSLCert}
	a.ListenInterface = c.Web.API.Interface + ":" + strconv.Itoa(c.Web.API.Port)
	return a
}

// Run starts the API server as specified in the configuration
func (a *API) Run() error {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{log.StandardLogger()}

	apiV1 := e.Group("/api/v1", middleware.CORS())
	apiV1.GET("/status", a.getStatus)
	apiV1.GET("/chains/:type/:hash", a.getBlock)
	apiV1.POST("/chains/:type", a.addBlock)
	apiV1.GET("/chains/:type", a.getBlocks)
	apiV1.GET("/search", a.getSearch)
	log.Infof("Starting API Server on interface %s", a.ListenInterface)
	return e.StartTLS(a.ListenInterface, a.certfile, a.keyfile)
}

func (a *API) getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, a.node.Status())
}

func (a *API) getBlock(c echo.Context) error {
	rh, err := base64.URLEncoding.DecodeString(c.Param("hash"))
	if err != nil {
		return err
	}
	var h [32]byte
	copy(h[:], rh)
	var b *chain.Block
	switch c.Param("type") {
	case "post":
		b = a.node.PostChain.Get(h)
	case "image":
		b = a.node.ImageChain.Get(h)
	case "key":
		b = a.node.KeyChain.Get(h)
	default:
		return c.JSON(http.StatusBadRequest, Error{Message: "Invalid Chain Type", Code: http.StatusBadRequest})
	}

	if b == nil {
		return c.JSON(http.StatusNotFound, Error{Message: "Block not found", Code: http.StatusNotFound})
	}
	return c.JSON(http.StatusOK, jsonize(b))
}

func (a *API) addBlock(c echo.Context) error {
	a.node.SubmitBlock(chain.Block{})
	return c.NoContent(http.StatusCreated)
}

func (a *API) getSearch(c echo.Context) error {
	results := []jsonBlock{}
	bs := a.node.PostChain.Search(c.QueryParam("q"))
	if len(bs) == 0 {
		return c.JSON(http.StatusNotFound, Error{Message: "No results found", Code: http.StatusNotFound})
	}
	for _, b := range bs {
		results = append(results, jsonize(b))
	}
	return c.JSON(http.StatusOK, struct {
		Results []jsonBlock `json:"results"`
	}{Results: results})
}

func (a *API) getBlocks(c echo.Context) error {
	results := []jsonBlock{}
	switch c.Param("type") {
	case "key", "image":
		return c.JSON(http.StatusBadRequest, Error{Code: http.StatusBadRequest, Message: "This operation is only supported for type=post"})
	case "post":
		for _, b := range a.node.ImageChain.Latest(10) {
			results = append(results, jsonize(b))
		}
	}
	return c.JSON(http.StatusOK, struct {
		Results []jsonBlock `json:"results"`
	}{Results: results})
}

func jsonize(b *chain.Block) jsonBlock {
	h := b.Hash()
	p := b.PrevHash
	return jsonBlock{
		Hash:      base64.URLEncoding.EncodeToString(h[:]),
		Nonce:     b.Nonce,
		PrevHash:  base64.URLEncoding.EncodeToString(p[:]),
		Content:   b.Content,
		Signature: b.Signature,
		Type:      b.Type,
		PubKey:    b.PubKey,
		Date:      b.Date.Unix(),
	}
}
