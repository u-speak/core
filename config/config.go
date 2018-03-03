package config

// Configuration is the exportable type of the node configuration
type Configuration struct {
	Version string
	Logger  struct {
		Format string `default:"default"`
		Debug  bool   `default:"false"`
	}
	Global struct {
		SSLCert string
		SSLKey  string
		Message string `default:"a nice person"`
		DNS     string `default:"discovery.uspeak.io"`
	}
	Storage struct {
		DataPath   string `default:"/var/lib/uspeak/data.db" env:"DATA_PATH"`
		TanglePath string `default:"/var/lib/uspeak/tangle.db" env:"TANGLE_PATH"`
	}
	NodeNetwork struct {
		Port      int    `default:"6969" env:"NODE_PORT"`
		Interface string `default:"127.0.0.1" env:"NODE_INTERFACE"`
	}
	Diagnostics struct {
		Port      int    `default:"1337" env:"DIAG_PORT"`
		Interface string `default:"127.0.0.1" env:"DIAG_INTERFACE"`
	}
	Hooks struct {
		PreAdd string
	}
	Web struct {
		Static struct {
			Port      int    `default:"4000" env:"WEB_PORT"`
			Interface string `default:"127.0.0.1" env:"WEB_INTERFACE"`
			Directory string `default:"portal/dist" env:"STATIC_DIR"`
		}
		MinUI struct {
			Enabled   bool   `default:"true"`
			Port      int    `default:"8080" env:"MINUI_PORT"`
			Interface string `default:"0.0.0.0" env:"MINUI_INTERFACE"`
		}
		API struct {
			Port           int    `default:"3000" env:"API_PORT"`
			Interface      string `default:"127.0.0.1"`
			PublicEndpoint string
			AdminEnabled   bool   `default:"false"`
			AdminUser      string `default:"admin"`
			AdminPassword  string `default:"admin"`
		}
	}
}
