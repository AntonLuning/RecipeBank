package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AntonLuning/RecipeBank/internal/core/storage"
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type RecipeService struct {
	storage storage.RecipeRepository
}

func NewRecipeService(storage storage.RecipeRepository) *RecipeService {
	return &RecipeService{
		storage: storage,
	}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	recipe, err := s.storage.GetRecipeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}
	return recipe, nil
}

func (s *RecipeService) GetRecipes(ctx context.Context, filter models.RecipeFilter, page, limit int) (*models.RecipePage, error) {
	if page < 1 || limit < 1 {
		return nil, fmt.Errorf("%w: page and limit must be positive", ErrInvalidInput)
	}

	recipes, err := s.storage.GetRecipes(ctx, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}
	return recipes, nil
}

func (s *RecipeService) CreateRecipe(ctx context.Context, recipe *models.Recipe) (*models.Recipe, error) {
	if err := validateRecipe(recipe); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	recipe.CreatedAt = time.Now()
	recipe.UpdatedAt = recipe.CreatedAt

	createdRecipe, err := s.storage.CreateRecipe(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}
	return createdRecipe, nil
}

func (s *RecipeService) UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error) {
	if err := validateRecipe(recipe); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	recipe.UpdatedAt = time.Now()

	updatedRecipe, err := s.storage.UpdateRecipe(ctx, id, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}
	return updatedRecipe, nil
}

func (s *RecipeService) DeleteRecipe(ctx context.Context, id string) error {
	if err := s.storage.DeleteRecipe(ctx, id); err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	return nil
}

// validateRecipe performs business logic validation on recipe data
func validateRecipe(recipe *models.Recipe) error {
	if recipe == nil {
		return fmt.Errorf("recipe cannot be nil")
	}

	if recipe.Title == "" {
		return fmt.Errorf("recipe title is required")
	}

	if len(recipe.Ingredients) == 0 {
		return fmt.Errorf("recipe must have at least one ingredient")
	}

	for i, ingredient := range recipe.Ingredients {
		if ingredient.Name == "" {
			return fmt.Errorf("ingredient %d must have a name", i+1)
		}
		if ingredient.Quantity <= 0 {
			return fmt.Errorf("ingredient %d must have a positive quantity", i+1)
		}
	}

	if len(recipe.Steps) == 0 {
		return fmt.Errorf("recipe must have at least one step")
	}

	for i, step := range recipe.Steps {
		if step == "" {
			return fmt.Errorf("step %d cannot be empty", i+1)
		}
	}

	if recipe.CookTime <= 0 {
		return fmt.Errorf("cook time must be positive")
	}

	if recipe.Servings <= 0 {
		return fmt.Errorf("servings must be positive")
	}

	return nil
}
