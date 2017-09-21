package webserver

import (
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/config"
	"net/http"
	"strconv"
)

// Server is a static webserver configured to log using the default logging methods
type Server struct {
	Directory string
	Interface string
}

// New returns a configured instance of the Webserver
func New(config config.Configuration) *Server {
	return &Server{
		Directory: config.Web.Static.Directory,
		Interface: config.Web.Static.Interface + ":" + strconv.Itoa(config.Web.Static.Port),
	}
}

// Run starts the server on the specified Port
func (s *Server) Run() {
	fs := http.FileServer(http.Dir(s.Directory))
	http.Handle("/", fs)

	log.Infof("Starting static webserver with directory %s", s.Directory)
	log.Fatal(http.ListenAndServe(s.Interface, nil))
}
