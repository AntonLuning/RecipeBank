package main

import (
	"context"
	"log/slog"

	"github.com/AntonLuning/RecipeBank/internal/core"
	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := storage.StorageConfig{
		Host:     "localhost",
		Port:     27017,
		Username: "root",
		Password: "example",
		Database: "recipes_db",
	} // TODO: use github.com/caarlos0/env/v11 for configuration parameters

	// Create storage
	storage, err := storage.NewMongoStorage(ctx, dbConfig)
	if err != nil {
		slog.Error("Unable to create new storage", "error", err.Error())
		return
	}

	// Ensure storage is closed when the program exits
	defer func() {
		if err := storage.Close(context.Background()); err != nil {
			slog.Error("Unable to close storage", "error", err.Error())
		}
	}()

	// Initialize storage (create indexes, etc.)
	if err := storage.Initialize(ctx); err != nil {
		slog.Error("Unable to initialize storage", "error", err.Error())
		return
	}

	// Initialize service layer
	recipeService := service.NewRecipeService(storage)

	// Initialize API server
	serverAddr := ":7777"
	server := core.NewAPIServer(serverAddr, recipeService)

	// Start the server
	if err := server.Run(); err != nil {
		slog.Error("Unable to run API server", "error", err.Error())
	}
}
