package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/views"
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
	"github.com/a-h/templ"
)

func GetIndexPage(w http.ResponseWriter, r *http.Request) {
	// Fetch recipes from the API
	resp, err := http.Get("http://localhost:9876/api/v1/recipe")
	if err != nil {
		slog.Error("Failed to fetch recipes", "error", err)
		http.Error(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the API response
	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		slog.Error("Failed to decode API response", "error", err)
		http.Error(w, "Failed to decode API response", http.StatusInternalServerError)
		return
	}

	// If the API request was not successful, return an error
	if !apiResp.Success {
		errorMsg := "Unknown error occurred"
		if apiResp.Error != nil {
			errorMsg = apiResp.Error.Message
		}
		slog.Error("API returned an error", "error", errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

	// Extract recipes from the API response
	var recipes []models.Recipe

	// Try to convert data to RecipePage first (most common response format)
	if recipePage, ok := apiResp.Data.(map[string]interface{}); ok {
		// Check if we have a recipes array in the data
		if recipesData, ok := recipePage["recipes"]; ok {
			// Convert to JSON and back to unmarshal into our slice
			recipesJSON, err := json.Marshal(recipesData)
			if err != nil {
				slog.Error("Failed to marshal recipes data", "error", err)
				http.Error(w, "Failed to process recipes", http.StatusInternalServerError)
				return
			}
			if err := json.Unmarshal(recipesJSON, &recipes); err != nil {
				slog.Error("Failed to parse recipes from page data", "error", err)
				http.Error(w, "Failed to process recipes", http.StatusInternalServerError)
				return
			}
			slog.Info("Successfully extracted recipes from RecipePage", "count", len(recipes))
		}
	}

	// If recipes is still empty, try to unmarshal directly as a recipe array
	if len(recipes) == 0 {
		recipesJSON, err := json.Marshal(apiResp.Data)
		if err != nil {
			slog.Error("Failed to marshal API data", "error", err)
			http.Error(w, "Failed to process recipes", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(recipesJSON, &recipes); err != nil {
			slog.Error("Failed to unmarshal recipes directly", "error", err, "data", string(recipesJSON))
			http.Error(w, "Failed to process recipes", http.StatusInternalServerError)
			return
		}
		slog.Info("Successfully extracted recipes directly from data", "count", len(recipes))
	}

	// Render the appropriate view based on request type
	if r.Header.Get("HX-Request") == "true" {
		// For HTMX requests, just render the recipe list component
		views.RecipeList(recipes).Render(r.Context(), w)
		return
	}

	if len(recipes) == 0 {
		slog.Warn("No recipes found to display")
	} else {
		slog.Info("Rendering recipes", "count", len(recipes), "first_recipe", recipes[0].Title)
	}

	// For normal requests, render the full page with layout
	recipeList := views.RecipeList(recipes)
	component := views.Layout("Recipes", recipeList)

	templ.Handler(component).ServeHTTP(w, r)
}
