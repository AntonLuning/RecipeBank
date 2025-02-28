package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AntonLuning/RecipeBank/pkg/core/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockStorage is a mock implementation of the storage.RecipeRepository interface
type MockStorage struct {
	mock.Mock
}

// GetRecipeByID mocks the GetRecipeByID method
func (m *MockStorage) GetRecipeByID(ctx context.Context, id string) (*models.Recipe, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

// GetRecipes mocks the GetRecipes method
func (m *MockStorage) GetRecipes(ctx context.Context, filter models.RecipeFilter, page, limit int) (*models.RecipePage, error) {
	args := m.Called(ctx, filter, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipePage), args.Error(1)
}

// CreateRecipe mocks the CreateRecipe method
func (m *MockStorage) CreateRecipe(ctx context.Context, recipe *models.Recipe) (*models.Recipe, error) {
	args := m.Called(ctx, recipe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

// UpdateRecipe mocks the UpdateRecipe method
func (m *MockStorage) UpdateRecipe(ctx context.Context, id string, recipe *models.Recipe) (*models.Recipe, error) {
	args := m.Called(ctx, id, recipe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

// DeleteRecipe mocks the DeleteRecipe method
func (m *MockStorage) DeleteRecipe(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Initialize mocks the Initialize method
func (m *MockStorage) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockStorage) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestGetRecipe tests the GetRecipe method
func TestGetRecipe(t *testing.T) {
	mockStorage := new(MockStorage)
	recipeService := NewRecipeService(mockStorage)

	ctx := context.Background()
	recipeID := "507f1f77bcf86cd799439011"
	objID, _ := primitive.ObjectIDFromHex(recipeID)

	t.Run("Success", func(t *testing.T) {
		expectedRecipe := &models.Recipe{
			ID:          objID,
			Title:       "Test Recipe",
			Description: "Test Description",
			Ingredients: []models.Ingredient{
				{Name: "Test Ingredient", Quantity: 1, Unit: "cup"},
			},
			Steps:     []string{"Step 1", "Step 2"},
			CookTime:  30,
			Servings:  4,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockStorage.On("GetRecipeByID", ctx, recipeID).Return(expectedRecipe, nil).Once()

		recipe, err := recipeService.GetRecipe(ctx, recipeID)

		assert.NoError(t, err)
		assert.Equal(t, expectedRecipe, recipe)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockStorage.On("GetRecipeByID", ctx, recipeID).Return(nil, expectedErr).Once()

		recipe, err := recipeService.GetRecipe(ctx, recipeID)

		assert.Error(t, err)
		assert.Nil(t, recipe)
		assert.Contains(t, err.Error(), "failed to get recipe")
		mockStorage.AssertExpectations(t)
	})
}

// TestGetRecipes tests the GetRecipes method
func TestGetRecipes(t *testing.T) {
	mockStorage := new(MockStorage)
	recipeService := NewRecipeService(mockStorage)

	ctx := context.Background()
	filter := models.RecipeFilter{
		Title: "Test",
	}

	t.Run("Success", func(t *testing.T) {
		expectedPage := &models.RecipePage{
			Recipes: []models.Recipe{
				{
					Title:       "Test Recipe",
					Description: "Test Description",
					Ingredients: []models.Ingredient{
						{Name: "Test Ingredient", Quantity: 1, Unit: "cup"},
					},
					Steps:     []string{"Step 1", "Step 2"},
					CookTime:  30,
					Servings:  4,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Total:      1,
			Page:       1,
			Limit:      10,
			TotalPages: 1,
		}

		mockStorage.On("GetRecipes", ctx, filter, 1, 10).Return(expectedPage, nil).Once()

		page, err := recipeService.GetRecipes(ctx, filter, 1, 10)

		assert.NoError(t, err)
		assert.Equal(t, expectedPage, page)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Invalid Page or Limit", func(t *testing.T) {
		page, err := recipeService.GetRecipes(ctx, filter, 0, 0)

		assert.Error(t, err)
		assert.Nil(t, page)
		assert.ErrorIs(t, errors.Unwrap(err), ErrInvalidInput)
	})

	t.Run("Storage Error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockStorage.On("GetRecipes", ctx, filter, 1, 10).Return(nil, expectedErr).Once()

		page, err := recipeService.GetRecipes(ctx, filter, 1, 10)

		assert.Error(t, err)
		assert.Nil(t, page)
		assert.Contains(t, err.Error(), "failed to get recipes")
		mockStorage.AssertExpectations(t)
	})
}

// TestCreateRecipe tests the CreateRecipe method
func TestCreateRecipe(t *testing.T) {
	mockStorage := new(MockStorage)
	recipeService := NewRecipeService(mockStorage)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		recipe := &models.Recipe{
			Title:       "Test Recipe",
			Description: "Test Description",
			Ingredients: []models.Ingredient{
				{Name: "Test Ingredient", Quantity: 1, Unit: "cup"},
			},
			Steps:    []string{"Step 1", "Step 2"},
			CookTime: 30,
			Servings: 4,
		}

		expectedRecipe := &models.Recipe{
			ID:          primitive.NewObjectID(),
			Title:       recipe.Title,
			Description: recipe.Description,
			Ingredients: recipe.Ingredients,
			Steps:       recipe.Steps,
			CookTime:    recipe.CookTime,
			Servings:    recipe.Servings,
		}

		mockStorage.On("CreateRecipe", ctx, mock.AnythingOfType("*models.Recipe")).Return(expectedRecipe, nil).Once()

		createdRecipe, err := recipeService.CreateRecipe(ctx, recipe)

		assert.NoError(t, err)
		assert.Equal(t, expectedRecipe, createdRecipe)
		assert.NotZero(t, recipe.CreatedAt)
		assert.NotZero(t, recipe.UpdatedAt)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		invalidRecipes := []*models.Recipe{
			nil,
			{Title: "", Ingredients: []models.Ingredient{{Name: "Test", Quantity: 1}}, Steps: []string{"Step 1"}, CookTime: 30, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{}, Steps: []string{"Step 1"}, CookTime: 30, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{{Name: "", Quantity: 1}}, Steps: []string{"Step 1"}, CookTime: 30, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{{Name: "Test", Quantity: 0}}, Steps: []string{"Step 1"}, CookTime: 30, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{{Name: "Test", Quantity: 1}}, Steps: []string{}, CookTime: 30, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{{Name: "Test", Quantity: 1}}, Steps: []string{""}, CookTime: 30, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{{Name: "Test", Quantity: 1}}, Steps: []string{"Step 1"}, CookTime: 0, Servings: 4},
			{Title: "Test", Ingredients: []models.Ingredient{{Name: "Test", Quantity: 1}}, Steps: []string{"Step 1"}, CookTime: 30, Servings: 0},
		}

		for _, invalidRecipe := range invalidRecipes {
			createdRecipe, err := recipeService.CreateRecipe(ctx, invalidRecipe)

			assert.Error(t, err)
			assert.Nil(t, createdRecipe)
			assert.ErrorIs(t, errors.Unwrap(err), ErrValidation)
		}
	})

	t.Run("Storage Error", func(t *testing.T) {
		recipe := &models.Recipe{
			Title:       "Test Recipe",
			Description: "Test Description",
			Ingredients: []models.Ingredient{
				{Name: "Test Ingredient", Quantity: 1, Unit: "cup"},
			},
			Steps:    []string{"Step 1", "Step 2"},
			CookTime: 30,
			Servings: 4,
		}

		expectedErr := errors.New("database error")
		mockStorage.On("CreateRecipe", ctx, mock.AnythingOfType("*models.Recipe")).Return(nil, expectedErr).Once()

		createdRecipe, err := recipeService.CreateRecipe(ctx, recipe)

		assert.Error(t, err)
		assert.Nil(t, createdRecipe)
		assert.Contains(t, err.Error(), "failed to create recipe")
		mockStorage.AssertExpectations(t)
	})
}

// TestUpdateRecipe tests the UpdateRecipe method
func TestUpdateRecipe(t *testing.T) {
	mockStorage := new(MockStorage)
	recipeService := NewRecipeService(mockStorage)

	ctx := context.Background()
	recipeID := "507f1f77bcf86cd799439011"

	t.Run("Success", func(t *testing.T) {
		recipe := &models.Recipe{
			Title:       "Updated Recipe",
			Description: "Updated Description",
			Ingredients: []models.Ingredient{
				{Name: "Updated Ingredient", Quantity: 2, Unit: "tbsp"},
			},
			Steps:    []string{"Updated Step 1", "Updated Step 2"},
			CookTime: 45,
			Servings: 6,
		}

		expectedRecipe := &models.Recipe{
			Title:       recipe.Title,
			Description: recipe.Description,
			Ingredients: recipe.Ingredients,
			Steps:       recipe.Steps,
			CookTime:    recipe.CookTime,
			Servings:    recipe.Servings,
		}

		mockStorage.On("UpdateRecipe", ctx, recipeID, mock.AnythingOfType("*models.Recipe")).Return(expectedRecipe, nil).Once()

		updatedRecipe, err := recipeService.UpdateRecipe(ctx, recipeID, recipe)

		assert.NoError(t, err)
		assert.Equal(t, expectedRecipe, updatedRecipe)
		assert.NotZero(t, recipe.UpdatedAt)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		invalidRecipe := &models.Recipe{
			Title:       "",
			Ingredients: []models.Ingredient{{Name: "Test", Quantity: 1}},
			Steps:       []string{"Step 1"},
			CookTime:    30,
			Servings:    4,
		}

		updatedRecipe, err := recipeService.UpdateRecipe(ctx, recipeID, invalidRecipe)

		assert.Error(t, err)
		assert.Nil(t, updatedRecipe)
		assert.ErrorIs(t, errors.Unwrap(err), ErrValidation)
	})

	t.Run("Storage Error", func(t *testing.T) {
		recipe := &models.Recipe{
			Title:       "Updated Recipe",
			Description: "Updated Description",
			Ingredients: []models.Ingredient{
				{Name: "Updated Ingredient", Quantity: 2, Unit: "tbsp"},
			},
			Steps:    []string{"Updated Step 1", "Updated Step 2"},
			CookTime: 45,
			Servings: 6,
		}

		expectedErr := errors.New("database error")
		mockStorage.On("UpdateRecipe", ctx, recipeID, mock.AnythingOfType("*models.Recipe")).Return(nil, expectedErr).Once()

		updatedRecipe, err := recipeService.UpdateRecipe(ctx, recipeID, recipe)

		assert.Error(t, err)
		assert.Nil(t, updatedRecipe)
		assert.Contains(t, err.Error(), "failed to update recipe")
		mockStorage.AssertExpectations(t)
	})
}

// TestDeleteRecipe tests the DeleteRecipe method
func TestDeleteRecipe(t *testing.T) {
	mockStorage := new(MockStorage)
	recipeService := NewRecipeService(mockStorage)

	ctx := context.Background()
	recipeID := "507f1f77bcf86cd799439011"

	t.Run("Success", func(t *testing.T) {
		mockStorage.On("DeleteRecipe", ctx, recipeID).Return(nil).Once()

		err := recipeService.DeleteRecipe(ctx, recipeID)

		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockStorage.On("DeleteRecipe", ctx, recipeID).Return(expectedErr).Once()

		err := recipeService.DeleteRecipe(ctx, recipeID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete recipe")
		mockStorage.AssertExpectations(t)
	})
}
