package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages/components"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetRecipesOverviewPartial(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Parse query parameters for search and pagination
		page := 1
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		limit := 12 // Default limit is 12 recipes per page
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
				limit = l
			}
		}

		// Build API endpoint with query parameters
		apiEndpoint := fmt.Sprintf("/recipe?page=%d&limit=%d", page, limit)

		// Add search filters if provided
		if title := r.URL.Query().Get("title"); title != "" {
			apiEndpoint += "&title=" + title
		}
		if ingredientNames := r.URL.Query().Get("ingredient_names"); ingredientNames != "" {
			apiEndpoint += "&ingredient_names=" + ingredientNames
		}
		if cookTime := r.URL.Query().Get("cook_time"); cookTime != "" {
			apiEndpoint += "&cook_time=" + cookTime
		}
		if tags := r.URL.Query().Get("tags"); tags != "" {
			apiEndpoint += "&tags=" + tags
		}

		// Fetch recipes from API
		recipePage, err := fetchRecipesFromAPI(apiURL, apiEndpoint)
		if err != nil {
			// Handle error - render page with error state
			component := components.RecipesOverview(components.RecipesPageData{
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
		component := components.RecipesOverview(components.RecipesPageData{
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

func fetchRecipesFromAPI(apiBaseURL string, apiEndpoint string) (*models.RecipePage, error) {
	recipePage, err := fetchFromAPI[models.RecipePage](apiBaseURL, apiEndpoint)
	time.Sleep(100 * time.Millisecond) // For user experience when perfroming lazy loading (avoiding flash of content)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}
	return recipePage, nil
}
