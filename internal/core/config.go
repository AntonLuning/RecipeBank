package core

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

var (
	instance *AppConfig
)

type AppConfig struct {
	// Applicaitons host name (IP)
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	// Applicaitons listen port
	Port uint16 `env:"PORT" envDefault:"9876"`
	// Database configuration (MongoDB)
	Database DatabaseConfig `envPrefix:"DB_"`
}

type DatabaseConfig struct {
	// Database host name (IP)
	Host string `env:"HOST,required"`
	// Database port
	Port uint16 `env:"PORT" envDefault:"27017"`
	// Database user name
	Username string `env:"USERNAME" envDefault:"root"`
	// Database password (for the given user name)
	Password string `env:"PASSWORD_FILE,required,file"`
	// Database name
	Database string `env:"DATABASE,required"`
}

func Config() AppConfig {
	if instance != nil {
		return *instance
	}

	opts := env.Options{
		Prefix: "RP_",
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
