package handlers

import (
	"encoding/json"
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

		// Build API request URL
		apiReqURL := fmt.Sprintf("%s/recipe/%s", apiURL, recipeID)

		// Fetch recipe from API
		recipe, err := fetchRecipeFromAPI(apiReqURL)
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

// fetchRecipeFromAPI fetches a single recipe from the API backend
func fetchRecipeFromAPI(apiURL string) (*models.Recipe, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipe: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("recipe not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResponse models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if !apiResponse.Success {
		errorMsg := "API request failed"
		if apiResponse.Error != nil {
			errorMsg = apiResponse.Error.Message
		}
		return nil, fmt.Errorf(errorMsg)
	}

	// The API response data should be a Recipe
	recipeData, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected API response format")
	}

	// Convert the map back to Recipe struct
	recipeJSON, err := json.Marshal(recipeData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipe data: %w", err)
	}

	var recipe models.Recipe
	if err := json.Unmarshal(recipeJSON, &recipe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe: %w", err)
	}

	return &recipe, nil
}
