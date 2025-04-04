package main

import (
	"log/slog"
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui"
)

func main() {
	cfg := ui.Config()

	// Initialize Server
	mux := http.NewServeMux()
	ui.InitAssets(mux, cfg.AssetsPath, cfg.Debug)
	ui.InitRoutes(mux)

	// Start the server
	if err := http.ListenAndServe(cfg.AppAddress(), mux); err != nil {
		slog.Error("Unable to run UI server", "error", err.Error())
	}
}
