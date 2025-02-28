package storage

import (
	"context"
	"testing"
	"time"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateRecipe(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	initTime := time.Now()

	tests := []struct {
		name    string
		recipe  *models.Recipe
		wantErr bool
	}{
		{
			name: "valid recipe",
			recipe: &models.Recipe{
				Title:       "Test Recipe",
				Description: "Test Description",
				Ingredients: []models.Ingredient{
					{Name: "ingredient1", Quantity: 1, Unit: "cup"},
				},
				Steps:     []string{"step1"},
				CookTime:  30,
				Servings:  4,
				Tags:      []string{"test"},
				CreatedAt: initTime,
				UpdatedAt: initTime,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.CreateRecipe(context.Background(), tt.recipe)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result.ID)
			assert.Equal(t, tt.recipe.Title, result.Title)
			assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Second)
			assert.Equal(t, result.CreatedAt, result.UpdatedAt)
		})
	}
}

func TestGetRecipeByID(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a test recipe
	recipe := &models.Recipe{
		Title:       "Test Recipe",
		Description: "Test Description",
		Ingredients: []models.Ingredient{
			{Name: "ingredient1", Quantity: 1, Unit: "cup"},
		},
		Steps:    []string{"step1"},
		CookTime: 30,
		Servings: 4,
		Tags:     []string{"test"},
	}

	created, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing recipe",
			id:      created.ID.Hex(),
			wantErr: false,
		},
		{
			name:    "non-existent recipe",
			id:      primitive.NewObjectID().Hex(),
			wantErr: true,
		},
		{
			name:    "invalid id format",
			id:      "invalid-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.GetRecipeByID(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, created.ID, result.ID)
			assert.Equal(t, created.Title, result.Title)
		})
	}
}

func TestUpdateRecipe(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	initTime := time.Now()

	// Create initial recipe
	recipe := &models.Recipe{
		Title:       "Original Recipe",
		Description: "Original Description",
		Ingredients: []models.Ingredient{
			{Name: "ingredient1", Quantity: 1, Unit: "cup"},
		},
		Steps:     []string{"step1"},
		CookTime:  30,
		Servings:  4,
		Tags:      []string{"test"},
		CreatedAt: initTime,
		UpdatedAt: initTime,
	}

	created, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	tests := []struct {
		name    string
		recipe  *models.Recipe
		wantErr bool
	}{
		{
			name: "valid update",
			recipe: &models.Recipe{
				ID:          created.ID,
				Title:       "Updated Recipe",
				Description: "Updated Description",
				Ingredients: []models.Ingredient{
					{Name: "ingredient2", Quantity: 2, Unit: "tbsp"},
				},
				Steps:     []string{"updated step"},
				CookTime:  45,
				Servings:  6,
				Tags:      []string{"updated"},
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "non-existent recipe",
			recipe: &models.Recipe{
				ID:          primitive.NewObjectID(),
				Title:       "Nonexistent",
				Description: "Should not update",
				Ingredients: []models.Ingredient{
					{Name: "ingredient", Quantity: 1, Unit: "cup"},
				},
				Steps: []string{"step"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.UpdateRecipe(context.Background(), tt.recipe.ID.Hex(), tt.recipe)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.recipe.Title, result.Title)
			assert.Equal(t, tt.recipe.Description, result.Description)
			assert.Equal(t, tt.recipe.Ingredients, result.Ingredients)
			assert.Equal(t, tt.recipe.Steps, result.Steps)
			assert.Equal(t, tt.recipe.CookTime, result.CookTime)
			assert.Equal(t, tt.recipe.Servings, result.Servings)
			assert.Equal(t, tt.recipe.Tags, result.Tags)
			assert.True(t, result.UpdatedAt.After(result.CreatedAt))
		})
	}
}

func TestDeleteRecipe(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a recipe to delete
	recipe := &models.Recipe{
		Title:       "Recipe to Delete",
		Description: "Will be deleted",
		Ingredients: []models.Ingredient{
			{Name: "ingredient1", Quantity: 1, Unit: "cup"},
		},
		Steps: []string{"step1"},
	}

	created, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing recipe",
			id:      created.ID.Hex(),
			wantErr: false,
		},
		{
			name:    "non-existent recipe",
			id:      primitive.NewObjectID().Hex(),
			wantErr: true,
		},
		{
			name:    "invalid id format",
			id:      "invalid-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DeleteRecipe(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify recipe is deleted
			_, err = storage.GetRecipeByID(context.Background(), tt.id)
			assert.Error(t, err)
		})
	}
}

func TestGetRecipes(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	initTime := time.Now()

	// Create test recipes
	recipes := []*models.Recipe{
		{
			Title:       "Pasta Carbonara",
			Description: "Italian classic",
			Ingredients: []models.Ingredient{
				{Name: "pasta", Quantity: 500, Unit: "g"},
				{Name: "eggs", Quantity: 3, Unit: "pieces"},
			},
			Steps:     []string{"step1", "step2"},
			CookTime:  20,
			Tags:      []string{"italian", "pasta"},
			CreatedAt: initTime,
			UpdatedAt: initTime,
		},
		{
			Title:       "Chicken Curry",
			Description: "Spicy curry",
			Ingredients: []models.Ingredient{
				{Name: "chicken", Quantity: 500, Unit: "g"},
				{Name: "curry powder", Quantity: 2, Unit: "tbsp"},
			},
			Steps:     []string{"step1", "step2"},
			CookTime:  45,
			Tags:      []string{"indian", "spicy"},
			CreatedAt: initTime,
			UpdatedAt: initTime,
		},
		{
			Title:       "Quick Pasta",
			Description: "Fast meal",
			Ingredients: []models.Ingredient{
				{Name: "pasta", Quantity: 250, Unit: "g"},
			},
			Steps:     []string{"step1"},
			CookTime:  15,
			Tags:      []string{"quick", "pasta"},
			CreatedAt: initTime,
			UpdatedAt: initTime,
		},
	}

	for _, r := range recipes {
		_, err := storage.CreateRecipe(context.Background(), r)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		filter        models.RecipeFilter
		page          int
		limit         int
		expectedCount int
	}{
		{
			name: "filter by title",
			filter: models.RecipeFilter{
				Title: "Pasta",
			},
			page:          1,
			limit:         10,
			expectedCount: 2,
		},
		{
			name: "filter by ingredient",
			filter: models.RecipeFilter{
				IngredientNames: []string{"chicken"},
			},
			page:          1,
			limit:         10,
			expectedCount: 1,
		},
		{
			name: "filter by cook time",
			filter: models.RecipeFilter{
				CookTime: 30,
			},
			page:          1,
			limit:         10,
			expectedCount: 2,
		},
		{
			name: "filter by tags",
			filter: models.RecipeFilter{
				Tags: []string{"pasta"},
			},
			page:          1,
			limit:         10,
			expectedCount: 2,
		},
		{
			name: "combined filters",
			filter: models.RecipeFilter{
				Title:           "Pasta",
				IngredientNames: []string{"pasta"},
				CookTime:        20,
				Tags:            []string{"quick"},
			},
			page:          1,
			limit:         10,
			expectedCount: 1,
		},
		{
			name:          "pagination - page 1",
			filter:        models.RecipeFilter{},
			page:          1,
			limit:         2,
			expectedCount: 2,
		},
		{
			name:          "pagination - page 2",
			filter:        models.RecipeFilter{},
			page:          2,
			limit:         2,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.GetRecipes(context.Background(), tt.filter, tt.page, tt.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(result.Recipes))

			// Verify pagination
			assert.Equal(t, tt.page, result.Page)
			assert.Equal(t, tt.limit, result.Limit)

			// Verify filters
			for _, recipe := range result.Recipes {
				if tt.filter.Title != "" {
					assert.Contains(t, recipe.Title, tt.filter.Title)
				}
				if len(tt.filter.Tags) > 0 {
					for _, tag := range tt.filter.Tags {
						assert.Contains(t, recipe.Tags, tag)
					}
				}
				if tt.filter.CookTime > 0 {
					assert.LessOrEqual(t, recipe.CookTime, tt.filter.CookTime)
				}
			}
		})
	}
}
