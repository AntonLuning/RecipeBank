package ai

import "github.com/AntonLuning/RecipeBank/pkg/core/models"

// RecipeAnalysisResult represents the structured output from the AI recipe analysis
type RecipeAnalysisResult struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Ingredients []models.Ingredient `json:"ingredients"`
	Steps       []string            `json:"steps"`
	CookTime    int                 `json:"cook_time"` // in minutes
	Servings    int                 `json:"servings"`
}

// JSONSchema returns a JSON schema example for the RecipeAnalysisResult
func (r *RecipeAnalysisResult) JSONSchema() string {
	return `{
  "title": "Spaghetti Carbonara",
  "description": "A classic Italian pasta dish with eggs, cheese, and pancetta",
  "ingredients": [
    {
      "name": "spaghetti",
      "quantity": 500,
      "unit": "g"
    },
    {
      "name": "salt"
    }
  ],
  "steps": [
    "Boil the pasta in salted water until al dente",
    "Fry the pancetta until crispy"
  ],
  "cook_time": 25,
  "servings": 4
}`
}
