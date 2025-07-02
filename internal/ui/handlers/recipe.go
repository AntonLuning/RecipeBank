package handlers

import (
	"fmt"
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetRecipePage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract recipe ID from URL
		recipeID := r.PathValue("id")
		if recipeID == "" {
			http.Error(w, "Recipe ID is required", http.StatusBadRequest)
			return
		}

		// Check if edit mode is requested
		editMode := r.URL.Query().Get("edit") == "true"

		// Fetch recipe from API
		recipe, err := fetchRecipeFromAPI(apiURL, recipeID)
		if err != nil {
			// Handle error - render page with error state
			component := pages.RecipePage(pages.RecipePageData{Error: err.Error()})
			renderComponent(w, r, component)
			return
		}

		// Render page with recipe
		component := pages.RecipePage(pages.RecipePageData{
			Recipe:   recipe,
			EditMode: editMode,
		})
		renderComponent(w, r, component)
	}
}

func UpdateRecipe(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract recipe ID from URL
		recipeID := r.PathValue("id")
		if recipeID == "" {
			http.Error(w, "Recipe ID is required", http.StatusBadRequest)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Get the current recipe first to merge changes
		currentRecipe, err := fetchRecipeFromAPI(apiURL, recipeID)
		if err != nil {
			http.Error(w, "Failed to fetch current recipe", http.StatusInternalServerError)
			return
		}

		// Update only the title for now
		if title := r.FormValue("title"); title != "" {
			currentRecipe.Title = title
		}

		// Send update to API
		recipe, err := updateRecipeViaAPI(apiURL, recipeID, currentRecipe)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update recipe: %v", err), http.StatusInternalServerError)
			return
		}

		// Render page with recipe
		component := pages.RecipePage(pages.RecipePageData{Recipe: recipe})
		renderComponent(w, r, component)
	}
}

func DeleteRecipe(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract recipe ID from URL
		recipeID := r.PathValue("id")
		if recipeID == "" {
			http.Error(w, "Recipe ID is required", http.StatusBadRequest)
			return
		}

		// Delete recipe via API
		err := deleteRecipeViaAPI(apiURL, recipeID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete recipe: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func fetchRecipeFromAPI(apiBaseURL string, recipeID string) (*models.Recipe, error) {
	apiEndpointWithID := fmt.Sprintf("/recipe/%s", recipeID)
	recipe, err := fetchFromAPI[models.Recipe](apiBaseURL, apiEndpointWithID)
	if err != nil {
		// Convert generic "resource not found" to more specific "recipe not found"
		if err.Error() == "resource not found" {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to fetch recipe: %w", err)
	}
	return recipe, nil
}

func updateRecipeViaAPI(apiBaseURL string, recipeID string, recipe *models.Recipe) (*models.Recipe, error) {
	apiEndpointWithID := fmt.Sprintf("/recipe/%s", recipeID)
	recipe, err := updateFromAPI(apiBaseURL, apiEndpointWithID, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}
	return recipe, nil
}

func deleteRecipeViaAPI(apiBaseURL string, recipeID string) error {
	apiEndpointWithID := fmt.Sprintf("/recipe/%s", recipeID)
	err := deleteFromAPI(apiBaseURL, apiEndpointWithID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	return nil
}
