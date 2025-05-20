package service

import (
	"context"
	"fmt"
	"time"

	"github.com/AntonLuning/RecipeBank/internal/core/ai"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type RecipeService struct {
	storage storage.RecipeStorage
	ai      ai.AI
}

func NewRecipeService(storage storage.RecipeStorage, ai ai.AI) *RecipeService {
	return &RecipeService{
		storage: storage,
		ai:      ai,
	}
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: invalid recipe ID", ErrInvalidInput)
	}

	recipe, err := s.storage.GetRecipeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}
	return recipe, nil
}

func (s *RecipeService) GetRecipes(ctx context.Context, filter models.RecipeFilter, page int, limit int) (*models.RecipePage, error) {
	// No validation here - storage layer handles default values

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

func (s *RecipeService) CreateRecipeFromImage(ctx context.Context, image string, imageType string) (*models.Recipe, error) {
	// Validate image
	if err := validateBase64Image(image, imageType); err != nil {
		return nil, fmt.Errorf("%w: image is not a valid %s (base64 encoded) or type is not supported", ErrValidation, imageType)
	}

	// Convert imageType to ImageContentType
	var imageContentType ai.ImageContentType
	switch imageType {
	case "jpeg", "jpg":
		imageContentType = ai.ImageContentTypeJPEG
	case "png":
		imageContentType = ai.ImageContentTypePNG
	default:
		return nil, fmt.Errorf("%w: image type %s is not supported", ErrValidation, imageType)
	}

	_, err := s.ai.AnalyzeImage(ctx, image, imageContentType, "recipe") // TODO: prompt and structured output
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe from image: %w", err)
	}

	// TODO: save recipe to storage

	return nil, nil
}

func (s *RecipeService) CreateRecipeFromURL(ctx context.Context, url string) (*models.Recipe, error) {
	// Validate URL and the it exists
	if err := validateURL(url); err != nil {
		return nil, fmt.Errorf("%w: URL could not be found: %s", ErrValidation, url)
	}

	_, err := s.ai.AnalyzeURL(ctx, url, "recipe") // TODO: prompt and structured output
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe from URL: %w", err)
	}

	// TODO: save recipe to storage

	return nil, nil
}

func (s *RecipeService) UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: invalid recipe ID", ErrInvalidInput)
	}

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
	if id == "" {
		return fmt.Errorf("%w: invalid recipe ID", ErrInvalidInput)
	}

	if err := s.storage.DeleteRecipe(ctx, id); err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	return nil
}

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
		if ingredient.Quantity < 0 {
			return fmt.Errorf("ingredient %s must have a positive quantity", ingredient.Name)
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

	if recipe.CookTime < 0 {
		return fmt.Errorf("cook time must be positive")
	}

	if recipe.Servings < 0 {
		return fmt.Errorf("servings must be positive")
	}

	return nil
}
