package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AntonLuning/RecipeBank/pkg/models"
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

// TestCreateRecipeWithDuplicateTitle tests creating a recipe with a title that already exists
func TestCreateRecipeWithDuplicateTitle(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a recipe
	recipe1 := &models.Recipe{
		Title:       "Duplicate Title Test",
		Description: "First recipe",
		Ingredients: []models.Ingredient{
			{Name: "ingredient1", Quantity: 1, Unit: "cup"},
		},
		Steps:    []string{"step1"},
		CookTime: 30,
		Servings: 4,
	}

	// Create the first recipe
	created1, err := storage.CreateRecipe(context.Background(), recipe1)
	require.NoError(t, err)
	require.NotNil(t, created1)

	// Try to create another recipe with the same title
	recipe2 := &models.Recipe{
		Title:       "Duplicate Title Test", // Same title
		Description: "Second recipe",
		Ingredients: []models.Ingredient{
			{Name: "ingredient2", Quantity: 2, Unit: "tbsp"},
		},
		Steps:    []string{"different step"},
		CookTime: 45,
		Servings: 2,
	}

	// This should still succeed as MongoDB doesn't enforce unique titles by default
	created2, err := storage.CreateRecipe(context.Background(), recipe2)
	require.NoError(t, err)
	require.NotNil(t, created2)

	// Verify they are different recipes
	assert.NotEqual(t, created1.ID, created2.ID)
	assert.Equal(t, created1.Title, created2.Title)
}

// TestGetRecipeByIDWithConcurrentAccess tests concurrent access to the same recipe
func TestGetRecipeByIDWithConcurrentAccess(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a test recipe
	recipe := &models.Recipe{
		Title:       "Concurrent Access Test",
		Description: "Test Description",
		Ingredients: []models.Ingredient{
			{Name: "ingredient1", Quantity: 1, Unit: "cup"},
		},
		Steps:    []string{"step1"},
		CookTime: 30,
		Servings: 4,
	}

	created, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	// Number of concurrent goroutines
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Create a channel to collect errors
	errChan := make(chan error, concurrency)

	// Launch multiple goroutines to access the same recipe
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			// Get the recipe
			retrieved, err := storage.GetRecipeByID(context.Background(), created.ID.Hex())
			if err != nil {
				errChan <- err
				return
			}

			// Verify the recipe
			if retrieved.ID != created.ID {
				errChan <- fmt.Errorf("ID mismatch: expected %v, got %v", created.ID, retrieved.ID)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check if there were any errors
	for err := range errChan {
		assert.NoError(t, err)
	}
}

// TestUpdateRecipeWithNonExistentID tests updating a recipe with a non-existent ID
func TestUpdateRecipeWithNonExistentID(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a non-existent ID
	nonExistentID := primitive.NewObjectID().Hex()

	// Create an update recipe
	updateRecipe := &models.Recipe{
		Title:       "Updated Recipe",
		Description: "Updated Description",
		Ingredients: []models.Ingredient{
			{Name: "updated ingredient", Quantity: 2, Unit: "tbsp"},
		},
		Steps:    []string{"updated step"},
		CookTime: 45,
		Servings: 6,
	}

	// Try to update a non-existent recipe
	_, err := storage.UpdateRecipe(context.Background(), nonExistentID, updateRecipe)

	// Should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestDeleteRecipeWithNonExistentID tests deleting a recipe with a non-existent ID
func TestDeleteRecipeWithNonExistentID(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a non-existent ID
	nonExistentID := primitive.NewObjectID().Hex()

	// Try to delete a non-existent recipe
	err := storage.DeleteRecipe(context.Background(), nonExistentID)

	// This should return an error since we're now checking if a document was actually deleted
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

// TestGetRecipesWithEmptyCollection tests getting recipes from an empty collection
func TestGetRecipesWithEmptyCollection(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Clear the collection to ensure it's empty
	err := storage.collection.Drop(context.Background())
	require.NoError(t, err)

	// Try to get recipes from an empty collection
	filter := models.RecipeFilter{}
	page, err := storage.GetRecipes(context.Background(), filter, 1, 10)

	// Should not return an error
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Empty(t, page.Recipes)
	assert.Equal(t, int64(0), page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.Limit)
	assert.Equal(t, 0, page.TotalPages)
}

// TestCreateRecipeWithLargeDocument tests creating a recipe with a very large document
func TestCreateRecipeWithLargeDocument(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a recipe with a large amount of data
	largeRecipe := &models.Recipe{
		Title:       "Large Recipe",
		Description: strings.Repeat("Very long description. ", 1000), // ~24KB of text
		Ingredients: make([]models.Ingredient, 100),
		Steps:       make([]string, 100),
		CookTime:    30,
		Servings:    4,
	}

	// Fill the ingredients and steps with data
	for i := 0; i < 100; i++ {
		largeRecipe.Ingredients[i] = models.Ingredient{
			Name:     fmt.Sprintf("Ingredient %d", i),
			Quantity: float32(i),
			Unit:     "unit",
		}
		largeRecipe.Steps[i] = fmt.Sprintf("Step %d: %s", i, strings.Repeat("Do something. ", 50))
	}

	// Try to create the large recipe
	created, err := storage.CreateRecipe(context.Background(), largeRecipe)

	// Should succeed as MongoDB can handle documents up to 16MB
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.ID)
}

// TestGetRecipesWithComplexFilter tests getting recipes with a complex filter
func TestGetRecipesWithComplexFilter(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create several recipes with different attributes
	recipes := []*models.Recipe{
		{
			Title:       "Pasta Carbonara",
			Description: "Italian pasta dish",
			Ingredients: []models.Ingredient{
				{Name: "pasta", Quantity: 200, Unit: "g"},
				{Name: "eggs", Quantity: 2, Unit: ""},
			},
			Steps:    []string{"step1", "step2"},
			CookTime: 20,
			Servings: 2,
			Tags:     []string{"italian", "pasta", "quick"},
		},
		{
			Title:       "Chicken Curry",
			Description: "Spicy chicken curry",
			Ingredients: []models.Ingredient{
				{Name: "chicken", Quantity: 500, Unit: "g"},
				{Name: "curry powder", Quantity: 2, Unit: "tbsp"},
			},
			Steps:    []string{"step1", "step2", "step3"},
			CookTime: 45,
			Servings: 4,
			Tags:     []string{"indian", "spicy", "chicken"},
		},
		{
			Title:       "Vegetable Stir Fry",
			Description: "Quick vegetable stir fry",
			Ingredients: []models.Ingredient{
				{Name: "mixed vegetables", Quantity: 400, Unit: "g"},
				{Name: "soy sauce", Quantity: 2, Unit: "tbsp"},
			},
			Steps:    []string{"step1", "step2"},
			CookTime: 15,
			Servings: 2,
			Tags:     []string{"vegetarian", "quick", "asian"},
		},
	}

	// Create all recipes
	for _, r := range recipes {
		_, err := storage.CreateRecipe(context.Background(), r)
		require.NoError(t, err)
	}

	// Test different filter combinations
	testCases := []struct {
		name          string
		filter        models.RecipeFilter
		expectedCount int
	}{
		{
			name:          "Filter by title containing 'Chicken'",
			filter:        models.RecipeFilter{Title: "Chicken"},
			expectedCount: 1,
		},
		{
			name:          "Filter by title containing 'quick' (case insensitive)",
			filter:        models.RecipeFilter{Title: "quick"},
			expectedCount: 0, // Title doesn't contain "quick", only the tags do
		},
		{
			name:          "Filter by tag 'quick'",
			filter:        models.RecipeFilter{Tags: []string{"quick"}},
			expectedCount: 2,
		},
		{
			name:          "Filter by multiple tags 'vegetarian' and 'quick'",
			filter:        models.RecipeFilter{Tags: []string{"vegetarian", "quick"}},
			expectedCount: 1,
		},
		{
			name:          "Filter by cook time",
			filter:        models.RecipeFilter{CookTime: 20},
			expectedCount: 2,
		},
		{
			name:          "Filter by title containing 'a'",
			filter:        models.RecipeFilter{Title: "a"},
			expectedCount: 2,
		},
		{
			name: "Complex filter: title containing 'a' AND tag 'quick'",
			filter: models.RecipeFilter{
				Title: "a",
				Tags:  []string{"quick"},
			},
			expectedCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page, err := storage.GetRecipes(context.Background(), tc.filter, 1, 10)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, len(page.Recipes), "Expected %d recipes, got %d", tc.expectedCount, len(page.Recipes))
		})
	}
}

func TestUpdateRecipeWithImage(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	initTime := time.Now()

	// Create initial recipe with image
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
		Image:     "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2Q==",
		CreatedAt: initTime,
		UpdatedAt: initTime,
	}

	created, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	// Update recipe with new image
	updateRecipe := &models.Recipe{
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
		Image:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAI9DeAQu3QAAAABJRU5ErkJggg=",
		UpdatedAt: time.Now(),
	}

	result, err := storage.UpdateRecipe(context.Background(), created.ID.Hex(), updateRecipe)
	require.NoError(t, err)

	// Verify the image was updated
	assert.Equal(t, updateRecipe.Image, result.Image)
	assert.Equal(t, updateRecipe.Title, result.Title)
	assert.Equal(t, updateRecipe.Description, result.Description)

	// Verify by retrieving the recipe from storage
	retrieved, err := storage.GetRecipeByID(context.Background(), created.ID.Hex())
	require.NoError(t, err)
	assert.Equal(t, updateRecipe.Image, retrieved.Image)
}

func TestGetIngredients(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create test recipes with various ingredients
	recipes := []*models.Recipe{
		{
			Title:       "Pasta Recipe",
			Description: "Simple pasta",
			Ingredients: []models.Ingredient{
				{Name: "Pasta", Quantity: 200, Unit: "g"},
				{Name: "Cheese", Quantity: 50, Unit: "g"},
			},
			Steps:    []string{"Cook pasta", "Add cheese"},
			CookTime: 15,
			Servings: 2,
		},
		{
			Title:       "Salad Recipe",
			Description: "Fresh salad",
			Ingredients: []models.Ingredient{
				{Name: "Lettuce", Quantity: 100, Unit: "g"},
				{Name: "Tomato", Quantity: 2, Unit: "pieces"},
				{Name: "Cheese", Quantity: 30, Unit: "g"}, // Cheese appears again
			},
			Steps:    []string{"Mix vegetables", "Add cheese"},
			CookTime: 5,
			Servings: 1,
		},
		{
			Title:       "Pizza Recipe",
			Description: "Homemade pizza",
			Ingredients: []models.Ingredient{
				{Name: "Flour", Quantity: 300, Unit: "g"},
				{Name: "Cheese", Quantity: 100, Unit: "g"}, // Cheese appears third time
				{Name: "Tomato", Quantity: 1, Unit: "can"}, // Tomato appears again
			},
			Steps:    []string{"Make dough", "Add toppings", "Bake"},
			CookTime: 30,
			Servings: 4,
		},
	}

	// Create all recipes
	for _, r := range recipes {
		_, err := storage.CreateRecipe(context.Background(), r)
		require.NoError(t, err)
	}

	testCases := []struct {
		name     string
		sort     string
		expected []models.ResourceSummary
	}{
		{
			name: "Sort by name ascending",
			sort: "name_asc",
			expected: []models.ResourceSummary{
				{Name: "Cheese", Count: 3},  // Appears in all 3 recipes
				{Name: "Flour", Count: 1},   // Appears in 1 recipe
				{Name: "Lettuce", Count: 1}, // Appears in 1 recipe
				{Name: "Pasta", Count: 1},   // Appears in 1 recipe
				{Name: "Tomato", Count: 2},  // Appears in 2 recipes
			},
		},
		{
			name: "Sort by name descending",
			sort: "name_desc",
			expected: []models.ResourceSummary{
				{Name: "Tomato", Count: 2},
				{Name: "Pasta", Count: 1},
				{Name: "Lettuce", Count: 1},
				{Name: "Flour", Count: 1},
				{Name: "Cheese", Count: 3},
			},
		},
		{
			name: "Sort by count ascending",
			sort: "count_asc",
			expected: []models.ResourceSummary{
				{Name: "Flour", Count: 1},
				{Name: "Lettuce", Count: 1},
				{Name: "Pasta", Count: 1},
				{Name: "Tomato", Count: 2},
				{Name: "Cheese", Count: 3},
			},
		},
		{
			name: "Sort by count descending",
			sort: "count_desc",
			expected: []models.ResourceSummary{
				{Name: "Cheese", Count: 3},
				{Name: "Tomato", Count: 2},
				{Name: "Flour", Count: 1}, // Count 1 items sorted alphabetically by name
				{Name: "Lettuce", Count: 1},
				{Name: "Pasta", Count: 1},
			},
		},
		{
			name: "Default sort (name_asc)",
			sort: "",
			expected: []models.ResourceSummary{
				{Name: "Cheese", Count: 3},
				{Name: "Flour", Count: 1},
				{Name: "Lettuce", Count: 1},
				{Name: "Pasta", Count: 1},
				{Name: "Tomato", Count: 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := storage.GetIngredients(context.Background(), tc.sort)
			require.NoError(t, err)

			// Verify the results match expectations
			require.Equal(t, len(tc.expected), len(result), "Expected %d ingredients, got %d", len(tc.expected), len(result))

			for i, expected := range tc.expected {
				assert.Equal(t, expected.Name, result[i].Name, "Ingredient %d name mismatch", i)
				assert.Equal(t, expected.Count, result[i].Count, "Ingredient %d count mismatch", i)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create test recipes with various tags
	recipes := []*models.Recipe{
		{
			Title:       "Pasta Recipe",
			Description: "Simple pasta",
			Ingredients: []models.Ingredient{
				{Name: "Pasta", Quantity: 200, Unit: "g"},
			},
			Steps:    []string{"Cook pasta"},
			CookTime: 15,
			Servings: 2,
			Tags:     []string{"italian", "quick", "vegetarian"},
		},
		{
			Title:       "Pizza Recipe",
			Description: "Homemade pizza",
			Ingredients: []models.Ingredient{
				{Name: "Flour", Quantity: 300, Unit: "g"},
			},
			Steps:    []string{"Make dough", "Bake"},
			CookTime: 30,
			Servings: 4,
			Tags:     []string{"italian", "baking"}, // italian appears again
		},
		{
			Title:       "Stir Fry Recipe",
			Description: "Quick stir fry",
			Ingredients: []models.Ingredient{
				{Name: "Vegetables", Quantity: 400, Unit: "g"},
			},
			Steps:    []string{"Stir fry"},
			CookTime: 10,
			Servings: 2,
			Tags:     []string{"quick", "asian", "vegetarian"}, // quick and vegetarian appear again
		},
	}

	// Create all recipes
	for _, r := range recipes {
		_, err := storage.CreateRecipe(context.Background(), r)
		require.NoError(t, err)
	}

	testCases := []struct {
		name     string
		sort     string
		expected []models.ResourceSummary
	}{
		{
			name: "Sort by name ascending",
			sort: "name_asc",
			expected: []models.ResourceSummary{
				{Name: "asian", Count: 1},      // Appears in 1 recipe
				{Name: "baking", Count: 1},     // Appears in 1 recipe
				{Name: "italian", Count: 2},    // Appears in 2 recipes
				{Name: "quick", Count: 2},      // Appears in 2 recipes
				{Name: "vegetarian", Count: 2}, // Appears in 2 recipes
			},
		},
		{
			name: "Sort by name descending",
			sort: "name_desc",
			expected: []models.ResourceSummary{
				{Name: "vegetarian", Count: 2},
				{Name: "quick", Count: 2},
				{Name: "italian", Count: 2},
				{Name: "baking", Count: 1},
				{Name: "asian", Count: 1},
			},
		},
		{
			name: "Sort by count ascending",
			sort: "count_asc",
			expected: []models.ResourceSummary{
				{Name: "asian", Count: 1},
				{Name: "baking", Count: 1},
				{Name: "italian", Count: 2},
				{Name: "quick", Count: 2},
				{Name: "vegetarian", Count: 2},
			},
		},
		{
			name: "Sort by count descending",
			sort: "count_desc",
			expected: []models.ResourceSummary{
				{Name: "italian", Count: 2},
				{Name: "quick", Count: 2},
				{Name: "vegetarian", Count: 2},
				{Name: "asian", Count: 1},
				{Name: "baking", Count: 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := storage.GetTags(context.Background(), tc.sort)
			require.NoError(t, err)

			// Verify the results match expectations
			require.Equal(t, len(tc.expected), len(result), "Expected %d tags, got %d", len(tc.expected), len(result))

			for i, expected := range tc.expected {
				assert.Equal(t, expected.Name, result[i].Name, "Tag %d name mismatch", i)
				assert.Equal(t, expected.Count, result[i].Count, "Tag %d count mismatch", i)
			}
		})
	}
}

func TestGetIngredientsEmptyCollection(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Test with empty collection
	result, err := storage.GetIngredients(context.Background(), "name_asc")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetTagsEmptyCollection(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Test with empty collection
	result, err := storage.GetTags(context.Background(), "name_asc")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetIngredientsWithNoIngredients(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a recipe with no ingredients (edge case)
	recipe := &models.Recipe{
		Title:       "No Ingredients Recipe",
		Description: "Strange recipe with no ingredients",
		Ingredients: []models.Ingredient{}, // Empty ingredients
		Steps:       []string{"Do nothing"},
		CookTime:    0,
		Servings:    1,
	}

	_, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	// Should return empty result since there are no ingredients
	result, err := storage.GetIngredients(context.Background(), "name_asc")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetTagsWithNoTags(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Create a recipe with no tags
	recipe := &models.Recipe{
		Title:       "No Tags Recipe",
		Description: "Recipe with no tags",
		Ingredients: []models.Ingredient{
			{Name: "Something", Quantity: 1, Unit: "unit"},
		},
		Steps:    []string{"Do something"},
		CookTime: 10,
		Servings: 1,
		Tags:     []string{}, // Empty tags
	}

	_, err := storage.CreateRecipe(context.Background(), recipe)
	require.NoError(t, err)

	// Should return empty result since there are no tags
	result, err := storage.GetTags(context.Background(), "name_asc")
	require.NoError(t, err)
	assert.Empty(t, result)
}
