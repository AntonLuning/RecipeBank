package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func CreateRecipeFromURL(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get the URL from form
		recipeURL := r.FormValue("url")
		if recipeURL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		// Create request payload
		requestPayload := models.CreateRecipeFromUrlRequest{
			URL: recipeURL,
		}

		// Call API to create recipe from URL
		recipe, err := createRecipeFromURLAPI(apiURL, requestPayload)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create recipe from URL: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to the recipe detail page in edit mode
		http.Redirect(w, r, fmt.Sprintf("/recipe/%s?edit=true", recipe.ID.Hex()), http.StatusSeeOther)
	}
}

func createRecipeFromURLAPI(apiBaseURL string, request models.CreateRecipeFromUrlRequest) (*models.Recipe, error) {
	apiEndpoint := "/recipe/ai/from-url"

	// Prepare request data
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	recipe, err := createFromAPI[models.Recipe](apiBaseURL, apiEndpoint, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe from URL: %w", err)
	}

	return recipe, nil
}
