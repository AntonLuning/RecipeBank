package ai

import (
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

// RecipeAnalysisResult represents the structured output from the AI recipe analysis
type RecipeAnalysisResult struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Ingredients []models.Ingredient `json:"ingredients"`
	Steps       []string            `json:"steps"`
	CookTime    int                 `json:"cook_time"` // in minutes
	Servings    int                 `json:"servings"`
}

// JSONSchema returns a JSON schema definition for the RecipeAnalysisResult
func (r *RecipeAnalysisResult) JSONSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]string{
				"type": "string",
			},
			"description": map[string]string{
				"type": "string",
			},
			"ingredients": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":     map[string]string{"type": "string"},
						"quantity": map[string]string{"type": "number"},
						"unit":     map[string]string{"type": "string"},
					},
					"required":             []string{"name", "quantity", "unit"},
					"additionalProperties": false,
				},
			},
			"steps": map[string]any{
				"type":  "array",
				"items": map[string]string{"type": "string"},
			},
			"cook_time": map[string]string{"type": "integer"},
			"servings":  map[string]string{"type": "integer"},
		},
		"required":             []string{"title", "description", "ingredients", "steps", "cook_time", "servings"},
		"additionalProperties": false,
	}
}
