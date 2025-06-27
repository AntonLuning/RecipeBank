package ui

import (
	"fmt"
	"strings"

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
	// API configuration
	API ApiConfig `envPrefix:"API_"`
}

type ApiConfig struct {
	// API URL (e.g., http://localhost:9876 or https://api.example.com)
	URL string `env:"URL,required"`
	// API base path (ignored if URL already contains a path)
	BasePath string `env:"BASE_PATH" envDefault:"/api/v1"`
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

func (c *AppConfig) ApiURL() string {
	url := strings.TrimSuffix(c.API.URL, "/")

	// If URL already contains a path, ignore BasePath
	if strings.Contains(strings.TrimPrefix(url, "http://"), "/") || strings.Contains(strings.TrimPrefix(url, "https://"), "/") {
		return url
	}

	basePath := strings.TrimPrefix(strings.TrimSuffix(c.API.BasePath, "/"), "/")
	if basePath == "" {
		return url
	}

	return fmt.Sprintf("%s/%s", url, basePath)
}
