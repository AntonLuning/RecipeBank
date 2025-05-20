package service

import (
	"context"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type Service interface {
	GetRecipe(ctx context.Context, id string) (*models.Recipe, error)
	GetRecipes(ctx context.Context, filter models.RecipeFilter, page int, limit int) (*models.RecipePage, error)
	CreateRecipe(ctx context.Context, recipe *models.Recipe) (*models.Recipe, error)
	CreateRecipeFromImage(ctx context.Context, image string, imageType string) (*models.Recipe, error)
	CreateRecipeFromURL(ctx context.Context, url string) (*models.Recipe, error)
	UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error)
	DeleteRecipe(ctx context.Context, id string) error
}
