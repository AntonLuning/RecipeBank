package ui

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

var (
	instance *AppConfig
)

type AppConfig struct {
	// Run the application with debug features (no-store static assets, etc.)
	Debug bool `env:"DEBUG" envDefault:"false"`
	// Applicaitons host name (IP)
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	// Applicaitons listen port
	Port uint16 `env:"PORT" envDefault:"9999"`
	// Path of static assets
	AssetsPath string `env:"ASSETS_PATH,required"`
}

func Config() AppConfig {
	if instance != nil {
		return *instance
	}

	opts := env.Options{
		Prefix: "RP_UI_",
	}

	config := AppConfig{}
	if err := env.ParseWithOptions(&config, opts); err != nil {
		panic(err.Error())
	}
	instance = &config

	return *instance
}

func (c *AppConfig) AppAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
