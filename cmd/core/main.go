package main

import (
	"log/slog"

	"github.com/AntonLuning/RecipeBank/internal/core"
)

func main() {
	server := core.NewApiServer(":7777")

	if err := server.Run(); err != nil {
		slog.Error("Unable to run API server", "error", err.Error())
	}
}
