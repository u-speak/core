package api

import (
	"encoding/base64"
	"errors"
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
	Message         string
	node            *node.Node
	certfile        string
	keyfile         string
	adminEnabled    bool
	user            string
	password        string
}

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type jsonBlock struct {
	Nonce     uint32 `json:"nonce"`
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
	a := &API{
		node:         n,
		keyfile:      c.Global.SSLKey,
		certfile:     c.Global.SSLCert,
		Message:      c.Global.Message,
		adminEnabled: c.Web.API.AdminEnabled,
		user:         c.Web.API.AdminUser,
		password:     c.Web.API.AdminPassword,
	}
	a.ListenInterface = c.Web.API.Interface + ":" + strconv.Itoa(c.Web.API.Port)
	return a
}

// Run starts the API server as specified in the configuration
func (a *API) Run() error {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{log.StandardLogger()}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:       middleware.DefaultSkipper,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		ExposeHeaders: []string{"X-Server-Message"},
	}))

	serverMessage := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Server-Message", a.Message)
			return next(c)
		}
	}

	e.Use(serverMessage)

	validateChain := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var ch *chain.Chain
			switch c.Param("type") {
			case "post":
				ch = a.node.PostChain
			case "image":
				ch = a.node.ImageChain
			case "key":
				ch = a.node.KeyChain
			}
			if !ch.Valid() {
				return c.JSON(http.StatusInternalServerError, Error{Code: http.StatusInternalServerError, Message: chain.ErrInvalidChain.Error()})
			}
			return next(c)
		}
	}

	apiV1 := e.Group("/api/v1")
	apiV1.GET("/status", a.getStatus)
	apiV1.GET("/chains/:type/:hash", a.getBlock, validateChain)
	apiV1.POST("/chains/:type", a.addBlock, validateChain)
	apiV1.GET("/chains/:type", a.getBlocks, validateChain)
	apiV1.GET("/search", a.getSearch)
	if a.adminEnabled {
		admin := apiV1.Group("/admin", middleware.BasicAuth(func(u, p string, c echo.Context) (bool, error) {
			if u == a.user && p == a.password {
				return true, nil
			}
			return false, nil
		}))
		admin.GET("/nodes", a.getNodes)
		admin.POST("/nodes", a.addNode)
		//admin.GET("/sync/:node", a.syncNode)
	}
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

func decodeHash(s string) ([32]byte, error) {
	h := [32]byte{}
	var hs []byte
	hs, err := base64.URLEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	hs, err = base64.StdEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	hs, err = base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	hs, err = base64.RawStdEncoding.DecodeString(s)
	if err == nil {
		copy(h[:], hs)
		return h, nil
	}
	return [32]byte{}, errors.New("Could not parse base64 data")
}

func (a *API) addBlock(c echo.Context) error {
	block := new(jsonBlock)
	if err := c.Bind(block); err != nil {
		return err
	}
	b := chain.Block{
		Content:   block.Content,
		Signature: block.Signature,
		Nonce:     block.Nonce,
		Type:      block.Type,
		Date:      time.Unix(block.Date, 0),
		PubKey:    block.PubKey,
	}
	hash, err := decodeHash(block.Hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
	}

	prevhash, err := decodeHash(block.PrevHash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
	}
	b.PrevHash = prevhash
	if b.Hash() != hash {
		h := b.Hash()
		log.Debugf("Should: %s, Was: %s", base64.URLEncoding.EncodeToString(h[:]), base64.URLEncoding.EncodeToString(hash[:]))
		return c.JSON(http.StatusBadRequest, Error{Code: http.StatusBadRequest, Message: "Block hash did not match its contents"})
	}
	a.node.SubmitBlock(b)
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
		l, err := a.node.PostChain.Latest(10)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Error{Code: http.StatusInternalServerError, Message: err.Error()})
		}
		for _, b := range l {
			results = append(results, jsonize(b))
		}
	}
	return c.JSON(http.StatusOK, struct {
		Results []jsonBlock `json:"results"`
	}{Results: results})
}

func (a *API) getNodes(c echo.Context) error {
	return c.JSON(http.StatusOK, a.node.Status().Connections)
}

func (a *API) addNode(c echo.Context) error {
	params := &struct {
		Node string `json:"node"`
	}{}
	if err := c.Bind(params); err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, Error{Code: http.StatusBadRequest, Message: "Bad request parameter"})
	}
	err := a.node.Connect(params.Node)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Error{
			Code:    http.StatusInternalServerError,
			Message: "Could not connect to remote: " + err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}

// func (a *API) syncNode(c echo.Context) error {
// 	//TODO: implement
// 	return nil
// }

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
