package service

import (
	"context"
	"fmt"

	"github.com/AntonLuning/RecipeBank/pkg/core"
)

type RecipeService struct {
}

func NewRecipeService() Service {
	return &RecipeService{}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string) (*core.Recipe, error) {
	recipe := core.Recipe{
		Title: fmt.Sprintf("Test with ID %s", id),
	}

	return &recipe, nil
}
