package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
	"github.com/AntonLuning/RecipeBank/internal/ui/pages/components"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetHomePage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		component := pages.HomePage()
		renderComponent(w, r, component)
	}
}

func GetRecipesOverviewPartial(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := createFilterFromQuery(r.URL.Query())

		// Fetch available resources from API (ingredients and tags) to be used in the filter
		filterData := components.FilterData{}
		availableIngredients, availableTags, err := fetchAvailableResourcesFromAPI(apiURL, r)
		if err != nil {
			// TODO: Log error
		} else {
			for _, ingredient := range availableIngredients.Resources {
				filterData.AvailableIngredients = append(filterData.AvailableIngredients, ingredient.Name)
			}
			for _, tag := range availableTags.Resources {
				filterData.AvailableTags = append(filterData.AvailableTags, tag.Name)
			}
		}

		recipePage, err := fetchRecipesFromAPI(apiURL, r)
		if err != nil {
			// Handle error - render partial with error state
			component := components.RecipesOverview(&filterData, &components.RecipesGridData{
				RecipePage: nil,
				Error:      err.Error(),
			}, &filter)
			renderComponent(w, r, component)
			return
		}

		// Render partial with recipes
		component := components.RecipesOverview(&filterData, &components.RecipesGridData{
			RecipePage: recipePage,
			Error:      "",
		}, &filter)
		renderComponent(w, r, component)
	}
}

func GetRecipesGridPartial(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := createFilterFromQuery(r.URL.Query())

		// Fetch recipes from API
		recipePage, err := fetchRecipesFromAPI(apiURL, r)
		if err != nil {
			// Handle error - render partial with error state
			component := components.RecipeGrid(&components.RecipesGridData{
				RecipePage: nil,
				Error:      err.Error(),
			}, &filter)
			renderComponent(w, r, component)
			return
		}

		// Render partial with recipes
		component := components.RecipeGrid(&components.RecipesGridData{
			RecipePage: recipePage,
			Error:      "",
		}, &filter)
		renderComponent(w, r, component)
	}
}

func fetchAvailableResourcesFromAPI(apiBaseURL string, r *http.Request) (*models.ResourcesResponse, *models.ResourcesResponse, error) {
	// Parse query parameters for search and pagination
	sort := "count_desc"
	if sortStr := r.URL.Query().Get("sort"); sortStr != "" {
		sort = url.QueryEscape(sortStr)
	}

	// Build API endpoint with query parameters
	ingredientsApiEndpoint := fmt.Sprintf("/recipe/ingredients?sort=%s", sort)
	tagsApiEndpoint := fmt.Sprintf("/recipe/tags?sort=%s", sort)

	// Use channels and goroutines for parallel execution
	type result struct {
		data *models.ResourcesResponse
		err  error
	}

	ingredientsChan := make(chan result, 1)
	tagsChan := make(chan result, 1)

	// Fetch ingredients in parallel
	go func() {
		data, err := fetchFromAPI[models.ResourcesResponse](apiBaseURL, ingredientsApiEndpoint)
		ingredientsChan <- result{data: data, err: err}
	}()

	// Fetch tags in parallel
	go func() {
		data, err := fetchFromAPI[models.ResourcesResponse](apiBaseURL, tagsApiEndpoint)
		tagsChan <- result{data: data, err: err}
	}()

	// Wait for both results
	ingredientsResult := <-ingredientsChan
	tagsResult := <-tagsChan

	if ingredientsResult.err != nil {
		return nil, nil, fmt.Errorf("failed to fetch ingredients: %w", ingredientsResult.err)
	}
	if tagsResult.err != nil {
		return nil, nil, fmt.Errorf("failed to fetch tags: %w", tagsResult.err)
	}

	return ingredientsResult.data, tagsResult.data, nil
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
