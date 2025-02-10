package service

import (
	"context"

	"github.com/AntonLuning/RecipeBank/pkg/core"
)

type Service interface {
	GetRecipe(ctx context.Context, id string) (*core.Recipe, error)
	// GetRecipes(context.Context, string) (*[]core.Recipe, error)
	// CreateRecipe(context.Context, string) error
	// UpdateRecipe(context.Context, string) error
}
