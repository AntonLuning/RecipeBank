package storage

import (
	"fmt"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type Storage struct {
	db []models.Recipe // TEMP
}

func NewStorage() Storage {
	return Storage{}
}

func (s *Storage) SaveRecipe(recipe models.Recipe) error {
	s.db = append(s.db, recipe)

	return nil
}

func (s *Storage) FetchRecipe(id string) (*models.Recipe, error) {
	for _, recipe := range s.db {
		if recipe.ID == id {
			return &recipe, nil
		}
	}

	return nil, fmt.Errorf("id %s could not be found", id)
}

func (s *Storage) FetchRecipes(filter string) (*[]models.Recipe, error) {
	// TODO: make use of filter

	return &s.db, nil
}
