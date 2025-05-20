package main

import (
	"context"
	"log/slog"

	"github.com/AntonLuning/RecipeBank/internal/core"
	"github.com/AntonLuning/RecipeBank/internal/core/ai"
	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := core.Config()

	dbConfig := storage.StorageConfig{
		Host:     cfg.Database.Host,
		Port:     int(cfg.Database.Port),
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
	}

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

	// Initialize AI client
	var aiClient ai.AI = nil
	switch cfg.AI.Provider {
	case "openai":
		aiClient = ai.NewOpenAI(cfg.AI.APIKey, cfg.AI.Model)
	default:
		slog.Warn("Empty or unsupported AI provider, running without AI", "provider", cfg.AI.Provider)
	}

	// Initialize service layer
	recipeService := service.NewRecipeService(storage, aiClient)

	// Initialize API server
	server := core.NewAPIServer(cfg.AppAddress(), recipeService)

	// Start the server
	if err := server.Run(); err != nil {
		slog.Error("Unable to run API server", "error", err.Error())
	}
}
