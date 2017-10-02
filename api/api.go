package api

import (
	"net/http"
	"strconv"

	"github.com/kpashka/echo-logrusmiddleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/node"
)

type API struct {
	ListenInterface string
}

// New returns a configured instance of the API server
func New(c config.Configuration) *API {
	a := &API{}
	a.ListenInterface = c.Web.API.Interface + ":" + strconv.Itoa(c.Web.API.Port)
	return a
}

func (a *API) Run() {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{logrus.StandardLogger()}

	apiV1 := e.Group("/api/v1", middleware.CORS())
	apiV1.GET("/status", a.getStatus)

	e.Logger.Error(e.Start(a.ListenInterface))
}

func (a *API) getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, &node.Status{Address: "example", Version: "1.0.0", Length: 0, Connections: 0})
}

//TODO: Add other handlers
