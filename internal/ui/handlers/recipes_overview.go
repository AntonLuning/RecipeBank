package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages/components"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

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
