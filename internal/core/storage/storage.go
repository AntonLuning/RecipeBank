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
	if recipe.Title == "" {
		return fmt.Errorf("title can not be empty")
	}

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

	if len(s.db) < 1 {
		return nil, fmt.Errorf("no recipe exists")
	}

	return &s.db, nil
}
