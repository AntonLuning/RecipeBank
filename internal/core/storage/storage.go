package storage

import (
	"context"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

// RecipeRepository defines the interface for recipe storage operations
type RecipeRepository interface {
	GetRecipeByID(ctx context.Context, id string) (*models.Recipe, error)
	GetRecipes(ctx context.Context, filter models.RecipeFilter, page, limit int) (*models.RecipePage, error)
	CreateRecipe(ctx context.Context, recipe *models.Recipe) (*models.Recipe, error)
	UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error)
	DeleteRecipe(ctx context.Context, id string) error
	Initialize(ctx context.Context) error
	Close(ctx context.Context) error
}
