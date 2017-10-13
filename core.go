package core

import (
	"github.com/u-speak/core/api"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/node"
	"github.com/u-speak/core/webserver"
)

// Config keeps the global configuration
var Config = config.Configuration{}

// Run Starts all core components
func Run() {
	n := node.New(Config)
	go n.Run()
	api.New(Config, n).Run()
}

// RunWeb starts a static webserver for the portal
func RunWeb() {
	webserver.New(Config).Run()
}
