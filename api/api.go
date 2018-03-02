package api

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"image/jpeg"
	"image/png"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/img"
	"github.com/u-speak/core/node"
	"github.com/u-speak/core/post"
	"github.com/u-speak/core/tangle"
	"github.com/u-speak/core/tangle/datastore"
	"github.com/u-speak/core/tangle/site"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/logrusmiddleware"
)

const (
	// MaxLatest is the highest limit amount for getRandom
	MaxLatest = 100
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

// Error is returned when something has gone wrong
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type jsonSite struct {
	Nonce        uint64                 `json:"nonce"`
	Validates    []string               `json:"validates"`
	Hash         string                 `json:"hash"`
	Content      string                 `json:"content"`
	Type         string                 `json:"type"`
	BubbleBabble string                 `json:"bubblebabble"`
	Weight       int                    `json:"weight"`
	Data         datastore.Serializable `json:"data"`
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

	apiV1 := e.Group("/api/v1")
	apiV1.GET("/status", a.getStatus)
	apiV1.POST("/image", a.uploadImage)
	apiV1.GET("/image/:hash", a.getImage)
	apiV1.GET("/tangle", a.getSearch)
	apiV1.GET("/tangle/random", a.getRandom)
	apiV1.GET("/tangle/:hash", a.getSite)
	apiV1.POST("/tangle/:hash", a.addSite)
	log.Infof("Starting API Server on interface %s", a.ListenInterface)
	return e.StartTLS(a.ListenInterface, a.certfile, a.keyfile)
}

func (a *API) getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, a.node.Status())
}

func (a *API) getSite(c echo.Context) error {
	h, err := decodeHash(c.Param("hash"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: "Invalid base64 data", Code: http.StatusBadRequest})
	}
	s := a.node.Tangle.Get(h)
	if s == nil {
		return c.JSON(http.StatusNotFound, Error{Message: "Site not found", Code: http.StatusNotFound})
	}
	err = s.Data.JSON()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Error{Message: "Error preparing response", Code: http.StatusInternalServerError})
	}
	j := JSONize(s)
	j.Weight = a.node.Tangle.Weight(s.Site)
	return c.JSON(http.StatusOK, j)
}

func (a *API) addSite(c echo.Context) error {
	s := new(jsonSite)
	switch c.Param("hash") {
	case "post":
		s.Data = &post.Post{}
	case "image":
		s.Data = &img.Image{}
	default:
		return c.JSON(http.StatusBadRequest, Error{Message: "Invalid type parameter: " + c.Param("hash"), Code: http.StatusInternalServerError})
	}
	if err := c.Bind(s); err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: err.Error(), Code: http.StatusBadRequest})
	}
	if err := s.Data.ReInit(); err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: err.Error(), Code: http.StatusBadRequest})
	}
	sh, err := decodeHash(s.Hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: "Could not decode provided hash", Code: http.StatusBadRequest})
	}
	switch c.Param("type") {
	case "post":
		err := verifyGPG(s.Data)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Error{Message: err.Error(), Code: http.StatusBadRequest})
		}
	}
	o := &tangle.Object{Data: s.Data}
	ch, err := decodeHash(s.Content)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: "Could not decode content hash", Code: http.StatusBadRequest})
	}
	dh, err := o.Data.Hash()
	if err != nil || ch != dh {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, Error{Message: "Content did not match supplied hash", Code: http.StatusBadRequest})
	}
	o.Site = &site.Site{Nonce: s.Nonce, Content: ch, Type: s.Type, Validates: []*site.Site{}}
	for _, b64 := range s.Validates {
		h, err := decodeHash(b64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Error{Message: "Invalid hash in validations: " + b64, Code: http.StatusBadRequest})
		}
		v := a.node.Tangle.Get(h)
		if v == nil {
			return c.JSON(http.StatusBadRequest, Error{Message: "Tried to verify unknown site " + b64, Code: http.StatusBadRequest})
		}
		o.Site.Validates = append(o.Site.Validates, v.Site)
	}
	if o.Site.Hash() != sh {
		return c.JSON(http.StatusBadRequest, Error{Message: "Provided hash does not match", Code: http.StatusBadRequest})
	}
	err = a.node.Submit(o)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: err.Error(), Code: http.StatusBadRequest})
	}
	return c.NoContent(http.StatusAccepted)
}

func (a *API) uploadImage(c echo.Context) error {
	o := &tangle.Object{Site: &site.Site{}}
	nonce, err := strconv.ParseUint(c.FormValue("nonce"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: err.Error(), Code: http.StatusBadRequest})
	}
	o.Site.Nonce = nonce
	o.Site.Type = "image"

	vls := strings.Split(c.FormValue("validates"), ",")
	for _, b64 := range vls {
		h, err := decodeHash(b64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Error{Message: "Invalid hash in validations: " + b64, Code: http.StatusBadRequest})
		}
		v := a.node.Tangle.Get(h)
		if v == nil {
			return c.JSON(http.StatusBadRequest, Error{Message: "Tried to verify unknown site " + b64, Code: http.StatusBadRequest})
		}
		o.Site.Validates = append(o.Site.Validates, v.Site)
	}
	rh, err := decodeHash(c.FormValue("hash"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: "Invalid field: Hash", Code: http.StatusBadRequest})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: "Could not find image", Code: http.StatusBadRequest})
	}
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: "Could not process image", Code: http.StatusBadRequest})
	}
	defer src.Close()

	buff := bytes.NewBuffer([]byte{})
	io.Copy(buff, src)
	if buff.Len() >= node.MaxMsgSize {
		return c.JSON(http.StatusBadRequest, Error{Message: "Image to large, please compress it further or crop it", Code: http.StatusBadRequest})
	}
	o.Data = &img.Image{Raw: buff.Bytes()}
	o.Site.Content, _ = o.Data.Hash()
	if o.Site.Hash() != rh {
		return c.JSON(http.StatusBadRequest, Error{Message: "Invalid hash. Please recalculate the nonce", Code: http.StatusBadRequest})
	}
	err = a.node.Submit(o)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Error{Message: err.Error(), Code: http.StatusBadRequest})
	}
	return c.NoContent(http.StatusAccepted)
}

func (a *API) getImage(c echo.Context) error {
	h, t := decodeImageHash(c.Param("hash"))
	s := a.node.Tangle.Get(h)
	if s.Site.Type != "image" {
		return c.JSON(http.StatusBadRequest, Error{Message: "requested site was not an image", Code: http.StatusBadRequest})
	}
	i, err := s.Data.(*img.Image).Image()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Error{Message: err.Error(), Code: http.StatusInternalServerError})
	}
	switch t {
	case "image/jpeg":
		c.Response().Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(c.Response().Writer, i, &jpeg.Options{Quality: 80})
		return nil
	case "image/png":
		c.Response().Header().Set("Content-Type", "image/png")
		png.Encode(c.Response().Writer, i)
		return nil
	default:
		return c.JSON(http.StatusBadRequest, Error{Message: "Please indicate the requested format with the Accept header or the file type", Code: http.StatusBadRequest})
	}
}

func (a *API) getSearch(c echo.Context) error {
	results := []jsonSite{}
	sr := a.node.Tangle.Search(c.QueryParam("q"))
	if len(sr) == 0 {
		return c.JSON(http.StatusNotFound, Error{Message: "No results found", Code: http.StatusNotFound})
	}
	for _, o := range sr {
		results = append(results, JSONize(o))
	}
	return c.JSON(http.StatusOK, struct {
		Results []jsonSite `json:"results"`
	}{Results: results})
}

func (a *API) getRandom(c echo.Context) error {
	ls := c.QueryParam("limit")
	limit := 10
	if ls != "" {
		ln, err := strconv.Atoi(ls)
		if err == nil && ln < MaxLatest {
			limit = ln
		}
	}
	hs := a.node.Tangle.Hashes()
	for i := range hs {
		j := rand.Intn(i + 1)
		hs[i], hs[j] = hs[j], hs[i]
	}
	res := []string{}
	if limit > len(hs) {
		limit = len(hs)
	}
	for _, h := range hs[:limit] {
		res = append(res, h.String())
	}
	return c.JSON(http.StatusOK, res)
}
