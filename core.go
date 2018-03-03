package core

import (
	"github.com/u-speak/core/api"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/diag"
	"github.com/u-speak/core/minui"
	"github.com/u-speak/core/node"
	"github.com/u-speak/core/webserver"

	log "github.com/sirupsen/logrus"
)

// Config keeps the global configuration
var Config = config.Configuration{}

// RunAPI starts the API server connected to the specific node
func RunAPI(n *node.Node) {
	err := api.New(Config, n).Run()
	if err != nil {
		log.Error(err)
	}
}

// RunDiag starts the diagnostics web interface
func RunDiag(n *node.Node) {
	err := diag.Run(Config, n)
	if err != nil {
		log.Error(err)
	}
}

// RunWeb starts a static webserver for the portal
func RunWeb() {
	webserver.New(Config).Run()
}

// RunMinUI starts the read-only minimal user interface for use on lower end devices
func RunMinUI(n *node.Node) {
	s := minui.New(Config, n)
	s.Run()
}
