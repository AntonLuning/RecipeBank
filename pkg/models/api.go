package models

// Request models

// @Description Query parameters for recipe search
type GetRecipesQuery struct {
	Page   int          `json:"page,omitempty" example:"1"`
	Limit  int          `json:"limit,omitempty" example:"10"`
	Filter RecipeFilter `json:"filter,omitempty"`
}

// @Description Recipe creation/update request
type RecipeRequest struct {
	Title       string       `json:"title" validate:"required" example:"Chocolate Chip Cookies"`
	Description string       `json:"description" example:"Delicious homemade chocolate chip cookies"`
	Ingredients []Ingredient `json:"ingredients" validate:"required,min=1,dive"`
	Steps       []string     `json:"steps" validate:"required,min=1" example:"['Preheat oven to 375°F', 'Mix ingredients', 'Bake for 10 minutes']"`
	CookTime    int          `json:"cook_time" example:"30"`
	Servings    int          `json:"servings" example:"12"`
	Tags        []string     `json:"tags" example:"['dessert', 'cookies', 'baking']"`
	Image       string       `json:"image,omitempty" example:"/9j/4AAQSkZJRgABAQAAAQ..."` // Base64 encoded image (optional)
}

// RecipeRequest aliases
type CreateRecipeRequest = RecipeRequest
type UpdateRecipeRequest = RecipeRequest

// @Description Request for AI-powered recipe creation from image
type CreateRecipeFromImageRequest struct {
	Image     string `json:"image" example:"/9j/4AAQSkZJRgABAQAAAQ..."` // Base64 encoded image
	ImageType string `json:"image_type" example:"jpeg"`                 // "jpeg", "jpg", "png"
}

// @Description Request for AI-powered recipe creation from URL
type CreateRecipeFromUrlRequest struct {
	URL string `json:"url" example:"https://example.com/recipe"` // URL to a webpage with recipe or to an image of a recipe
}

// @Description Query parameters for retrieving resources (ingredients, tags)
type GetResourcesQuery struct {
	Sort string `json:"sort,omitempty" example:"name_asc"` // Options: "name_asc", "name_desc", "count_asc", "count_desc"
}

// Response models

// @Description Standard API response wrapper
type APIResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// @Description Resources (ingredients, tags) list response
type ResourcesResponse struct {
	Resources []ResourceSummary `json:"resources"`
	Total     int               `json:"total" example:"42"` // Total number of unique resources
}

// @Description API error information
type APIError struct {
	Code    string `json:"code" example:"validation_error"`
	Message string `json:"message" example:"The provided input data is invalid"`
}
