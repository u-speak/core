package core

import (
	"github.com/u-speak/core/api"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/webserver"
)

// Config keeps the global configuration
var Config = config.Configuration{}

// RunWeb starts the static webserver
func RunWeb() {
	webserver.New(Config).Run()
}

// RunNode starts the node service
func RunNode() {
	for {
		//TODO: Implement this
	}
}

func RunAPI() {
	api.New(Config).Run()
}
