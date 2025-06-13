package handlers

import (
	"fmt"
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetRecipePage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Extract recipe ID from URL
		recipeID := r.PathValue("id")
		if recipeID == "" {
			http.Error(w, "Recipe ID is required", http.StatusBadRequest)
			return
		}

		// Fetch recipe from API
		recipe, err := fetchRecipeFromAPI(apiURL, "/recipe", recipeID)
		if err != nil {
			// Handle error - render page with error state
			component := pages.RecipePage(pages.RecipePageData{
				Recipe: nil,
				Error:  err.Error(),
			})
			if renderErr := component.Render(r.Context(), w); renderErr != nil {
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
			return
		}

		// Render page with recipe
		component := pages.RecipePage(pages.RecipePageData{
			Recipe: recipe,
			Error:  "",
		})
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
	}
}

func fetchRecipeFromAPI(apiBaseURL string, apiEndpoint string, recipeID string) (*models.Recipe, error) {
	apiEndpointWithID := fmt.Sprintf("%s/%s", apiEndpoint, recipeID)
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
