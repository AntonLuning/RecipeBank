package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages/components"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetRecipesGridPartial(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		filter := createFilterFromQuery(r.URL.Query())

		// Fetch recipes from API
		recipePage, err := fetchRecipesFromAPI(apiURL, r)
		if err != nil {
			// Handle error - render partial with error state
			component := components.RecipeGrid(&components.RecipesGridData{
				RecipePage: nil,
				Error:      err.Error(),
			}, &filter)
			if renderErr := component.Render(r.Context(), w); renderErr != nil {
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
			return
		}

		// Render partial with recipes
		component := components.RecipeGrid(&components.RecipesGridData{
			RecipePage: recipePage,
			Error:      "",
		}, &filter)
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
	}
}

func fetchRecipesFromAPI(apiBaseURL string, r *http.Request) (*models.RecipePage, error) {
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
		apiEndpoint += "&title=" + url.QueryEscape(title)
	}
	if ingredients := r.URL.Query().Get("ingredient_names"); ingredients != "" {
		apiEndpoint += "&ingredients=" + url.QueryEscape(ingredients)
	}
	if cookTime := r.URL.Query().Get("cook_time"); cookTime != "" {
		apiEndpoint += "&cook_time=" + url.QueryEscape(cookTime)
	}
	if tags := r.URL.Query().Get("tags"); tags != "" {
		apiEndpoint += "&tags=" + url.QueryEscape(tags)
	}

	recipePage, err := fetchFromAPI[models.RecipePage](apiBaseURL, apiEndpoint)
	time.Sleep(250 * time.Millisecond) // For user experience when perfroming lazy loading (avoiding flash of content)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}
	return recipePage, nil
}

func createFilterFromQuery(query url.Values) models.RecipeFilter {
	cookTime, err := strconv.Atoi(query.Get("cook_time"))
	if err != nil {
		cookTime = 0
	}

	return models.RecipeFilter{
		Title:           query.Get("title"),
		IngredientNames: parseStringArray(query.Get("ingredient_names")),
		CookTime:        cookTime,
		Tags:            parseStringArray(query.Get("tags")),
	}
}

func parseStringArray(input string) []string {
	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
