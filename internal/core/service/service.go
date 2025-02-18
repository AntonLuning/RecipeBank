package service

import (
	"context"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type Service interface {
	GetRecipe(ctx context.Context, id string) (recipe *models.Recipe, err error)
	GetRecipes(ctx context.Context, filter string) (recipes *[]models.Recipe, err error)
	CreateRecipe(ctx context.Context, recipe models.PostRecipeData) (id string, err error)
	// UpdateRecipe(context.Context, string) error
}
