package models

// Request models
type GetRecipesQuery struct {
	Page   int          `json:"page,omitempty"`
	Limit  int          `json:"limit,omitempty"`
	Filter RecipeFilter `json:"filter,omitempty"`
}

type RecipeRequest struct {
	Title       string       `json:"title" validate:"required"`
	Description string       `json:"description"`
	Ingredients []Ingredient `json:"ingredients" validate:"required,min=1,dive"`
	Steps       []string     `json:"steps" validate:"required,min=1"`
	CookTime    int          `json:"cook_time"`
	Servings    int          `json:"servings"`
	Tags        []string     `json:"tags"`
}

// Alias the RecipeRequest for better semantics
type CreateRecipeRequest = RecipeRequest
type UpdateRecipeRequest = RecipeRequest

// Response models
type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
