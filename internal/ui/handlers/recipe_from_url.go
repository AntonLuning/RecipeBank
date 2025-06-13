package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

		// Redirect to the recipe detail page
		http.Redirect(w, r, fmt.Sprintf("/recipe/%s", recipe.ID.Hex()), http.StatusSeeOther)
	}
}

func createRecipeFromURLAPI(apiBaseURL string, request models.CreateRecipeFromUrlRequest) (*models.Recipe, error) {
	// Prepare API request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API call
	apiURL := fmt.Sprintf("%s/recipe/ai/from-url", strings.TrimRight(apiBaseURL, "/"))
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse API response
	var apiResponse models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if !apiResponse.Success {
		errorMsg := "API request failed"
		if apiResponse.Error != nil {
			errorMsg = apiResponse.Error.Message
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// Convert response data to Recipe
	responseData, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected API response format")
	}

	dataJSON, err := json.Marshal(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	var recipe models.Recipe
	if err := json.Unmarshal(dataJSON, &recipe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe data: %w", err)
	}

	return &recipe, nil
}
