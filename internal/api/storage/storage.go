package storage

import (
	"context"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

// RecipeStorage defines the interface for recipe storage operations
type RecipeStorage interface {
	Initialize(ctx context.Context) error
	Close(ctx context.Context) error
	GetRecipeByID(ctx context.Context, id string) (*models.Recipe, error)
	GetRecipes(ctx context.Context, filter models.RecipeFilter, page, limit int) (*models.RecipePage, error)
	CreateRecipe(ctx context.Context, recipe *models.Recipe) (*models.Recipe, error)
	UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error)
	DeleteRecipe(ctx context.Context, id string) error
	GetIngredients(ctx context.Context, sort string) ([]models.ResourceSummary, error)
	GetTags(ctx context.Context, sort string) ([]models.ResourceSummary, error)
}
