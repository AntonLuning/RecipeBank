package main

import (
	"log/slog"

	"github.com/AntonLuning/RecipeBank/internal/core"
	"github.com/AntonLuning/RecipeBank/internal/core/service"
)

func main() {
	recipeService := service.NewRecipeService()
	// if cfg.InludeMetrics {
	// 	recipeService = service.NewMetricsService(recipeService, fmt.Sprintf(":%d", cfg.PortMetrics)) // Service wrapped in metrics
	// }

	server := core.NewApiServer(":7777", recipeService)

	if err := server.Run(); err != nil {
		slog.Error("Unable to run API server", "error", err.Error())
	}
}
