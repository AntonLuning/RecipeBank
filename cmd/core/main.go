package main

import (
	"log/slog"

	"github.com/AntonLuning/RecipeBank/internal/core"
	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
)

func main() {
	dbConfig := storage.StorageConfig{
		Host:     "localhost",
		Port:     27017,
		Username: "root",
		Password: "example",
		Database: "bank",
	} // TODO: use github.com/caarlos0/env/v11 for configuration parameters
	storage, err := storage.NewStorage(dbConfig)
	if err != nil {
		slog.Error("Unable to create new storage", "error", err.Error())
		return
	}

	recipeService := service.NewRecipeService(storage)

	server := core.NewApiServer(":7777", recipeService)

	if err := server.Run(); err != nil {
		slog.Error("Unable to run API server", "error", err.Error())
	}
}
