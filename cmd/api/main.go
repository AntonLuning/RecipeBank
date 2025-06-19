package main

import (
	"context"
	"log/slog"

	core "github.com/AntonLuning/RecipeBank/internal/api"
	"github.com/AntonLuning/RecipeBank/internal/api/ai"
	"github.com/AntonLuning/RecipeBank/internal/api/service"
	"github.com/AntonLuning/RecipeBank/internal/api/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := core.Config()

	// Create storage
	dbConfig := storage.StorageConfig{
		Host:     cfg.Database.Host,
		Port:     int(cfg.Database.Port),
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
	}
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
	var aiClient ai.RecipeAI = nil
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
