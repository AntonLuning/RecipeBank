package models

// Request models

// GetRecipesQuery represents query parameters for getting recipes
// @Description Query parameters for recipe search
type GetRecipesQuery struct {
	Page   int          `json:"page,omitempty" example:"1"`
	Limit  int          `json:"limit,omitempty" example:"10"`
	Filter RecipeFilter `json:"filter,omitempty"`
}

// RecipeRequest represents the request body for creating/updating recipes
// @Description Recipe creation/update request
type RecipeRequest struct {
	Title       string       `json:"title" validate:"required" example:"Chocolate Chip Cookies"`
	Description string       `json:"description" example:"Delicious homemade chocolate chip cookies"`
	Ingredients []Ingredient `json:"ingredients" validate:"required,min=1,dive"`
	Steps       []string     `json:"steps" validate:"required,min=1" example:"['Preheat oven to 375°F', 'Mix ingredients', 'Bake for 10 minutes']"`
	CookTime    int          `json:"cook_time" example:"30"`
	Servings    int          `json:"servings" example:"12"`
	Tags        []string     `json:"tags" example:"['dessert', 'cookies', 'baking']"`
}

// Alias the RecipeRequest for better semantics
type CreateRecipeRequest = RecipeRequest
type UpdateRecipeRequest = RecipeRequest

// CreateRecipeFromImageRequest represents the request for creating a recipe from an image
// @Description Request for AI-powered recipe creation from image
type CreateRecipeFromImageRequest struct {
	Image     string `json:"image" example:"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ..."` // Base64 encoded image
	ImageType string `json:"image_type" example:"jpeg"`                                        // "jpeg", "jpg", "png"
}

// CreateRecipeFromUrlRequest represents the request for creating a recipe from a URL
// @Description Request for AI-powered recipe creation from URL
type CreateRecipeFromUrlRequest struct {
	URL string `json:"url" example:"https://example.com/recipe"` // URL to a webpage with recipe or to an image of a recipe
}

// Response models

// APIResponse represents the standard API response format
// @Description Standard API response wrapper
type APIResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError represents an API error
// @Description API error information
type APIError struct {
	Code    string `json:"code" example:"validation_error"`
	Message string `json:"message" example:"The provided input data is invalid"`
}
