package api

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

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
	a := &API{node: n}
	a.ListenInterface = c.Web.API.Interface + ":" + strconv.Itoa(c.Web.API.Port)
	return a
}

// Run starts the API server as specified in the configuration
func (a *API) Run() {
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
	e.Logger.Error(e.Start(a.ListenInterface))
}

func (a *API) getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, &node.Status{Address: "example", Version: "1.0.0", Length: 0, Connections: 0})
}

func (a *API) getBlock(c echo.Context) error {
	return c.JSON(http.StatusOK, jsonize(&chain.Block{Date: time.Now()}))
}

func (a *API) addBlock(c echo.Context) error {
	a.node.SubmitBlock(chain.Block{})
	return c.NoContent(http.StatusCreated)
}

func (a *API) getSearch(c echo.Context) error {
	results := []jsonBlock{}
	for i := 0; i < 50; i++ {
		results = append(results, jsonize(&chain.Block{Nonce: 42, Date: time.Now(), Content: "Result" + strconv.Itoa(i)}))
	}
	return c.JSON(http.StatusOK, struct {
		Results []jsonBlock `json:"results"`
	}{Results: results})
}

func (a *API) getBlocks(c echo.Context) error {
	results := []jsonBlock{}
	for i := 0; i < 50; i++ {
		results = append(results, jsonize(&chain.Block{Nonce: 42, Date: time.Now(), Content: "Block" + strconv.Itoa(i)}))
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
