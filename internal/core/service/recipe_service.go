package service

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/AntonLuning/RecipeBank/internal/core/storage"
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type RecipeService struct {
	storage *storage.Storage
}

func NewRecipeService(storage *storage.Storage) Service {
	return &RecipeService{
		storage: storage,
	}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	return s.storage.FetchRecipe(id)
}

func (s *RecipeService) GetRecipes(ctx context.Context, filter string) (*[]models.Recipe, error) {
	return s.storage.FetchRecipes(filter)
}

func (s *RecipeService) CreateRecipe(ctx context.Context, recipeData models.PostRecipeData) (string, error) {
	id := strconv.Itoa(rand.Int())

	recipe := models.Recipe{
		ID:    id,
		Title: recipeData.Title,
	}

	if err := s.storage.SaveRecipe(recipe); err != nil {
		return "", err
	}

	return id, nil
}
