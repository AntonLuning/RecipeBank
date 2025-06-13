package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetIndexPage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Parse query parameters for search and pagination
		page := 1
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		limit := 12
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
				limit = l
			}
		}

		// Build API request URL with query parameters
		apiReqURL := fmt.Sprintf("%s/recipe?page=%d&limit=%d", apiURL, page, limit)

		// Add search filters if provided
		if title := r.URL.Query().Get("title"); title != "" {
			apiReqURL += "&title=" + title
		}
		if ingredientNames := r.URL.Query().Get("ingredient_names"); ingredientNames != "" {
			apiReqURL += "&ingredient_names=" + ingredientNames
		}
		if cookTime := r.URL.Query().Get("cook_time"); cookTime != "" {
			apiReqURL += "&cook_time=" + cookTime
		}
		if tags := r.URL.Query().Get("tags"); tags != "" {
			apiReqURL += "&tags=" + tags
		}

		// Fetch recipes from API
		recipePage, err := fetchRecipesFromAPI(apiReqURL)
		if err != nil {
			// Handle error - render page with error state
			component := pages.RecipesOverviewPage(pages.RecipesPageData{
				RecipePage:  nil,
				Error:       err.Error(),
				SearchQuery: r.URL.Query(),
			})
			if renderErr := component.Render(r.Context(), w); renderErr != nil {
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
			return
		}

		// Render page with recipes
		component := pages.RecipesOverviewPage(pages.RecipesPageData{
			RecipePage:  recipePage,
			Error:       "",
			SearchQuery: r.URL.Query(),
		})
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
	}
}

// fetchRecipesFromAPI fetches recipes from the API backend
func fetchRecipesFromAPI(apiURL string) (*models.RecipePage, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}
	defer resp.Body.Close()

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

	// The API response data should be a RecipePage
	recipePageData, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected API response format")
	}

	// Convert the map back to RecipePage struct
	recipePageJSON, err := json.Marshal(recipePageData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipe page data: %w", err)
	}

	var recipePage models.RecipePage
	if err := json.Unmarshal(recipePageJSON, &recipePage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipe page: %w", err)
	}

	return &recipePage, nil
}
