package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Recipe represents a recipe in the system
// @Description Recipe information
type Recipe struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"507f1f77bcf86cd799439011"`
	Title       string             `bson:"title" json:"title" example:"Chocolate Chip Cookies"`
	Description string             `bson:"description" json:"description" example:"Delicious homemade chocolate chip cookies"`
	Ingredients []Ingredient       `bson:"ingredients" json:"ingredients"`
	Steps       []string           `bson:"steps" json:"steps" example:"['Preheat oven to 375°F', 'Mix ingredients', 'Bake for 10 minutes']"`
	CookTime    int                `bson:"cook_time" json:"cook_time,omitempty" example:"30"` // in minutes
	Servings    int                `bson:"servings" json:"servings,omitempty" example:"12"`
	Tags        []string           `bson:"tags" json:"tags,omitempty" example:"['dessert', 'cookies', 'baking']"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at" example:"2023-01-15T09:30:00Z"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at" example:"2023-01-15T09:30:00Z"`
}

// Ingredient represents an ingredient in a recipe
// @Description Ingredient information
type Ingredient struct {
	Name     string  `bson:"name" json:"name" example:"Flour"`
	Quantity float32 `bson:"quantity,omitempty" json:"quantity,omitempty" example:"2.5"`
	Unit     string  `bson:"unit,omitempty" json:"unit,omitempty" example:"cups"`
}

// RecipeFilter represents filters for searching recipes
// @Description Filter criteria for searching recipes
type RecipeFilter struct {
	Title           string   `json:"title,omitempty" example:"Chocolate"`
	IngredientNames []string `json:"ingredient_names,omitempty" example:"['flour', 'sugar']"`
	CookTime        int      `json:"cook_time,omitempty" example:"30"`
	Tags            []string `json:"tags,omitempty" example:"['dessert', 'quick']"`
}

// RecipePage represents a paginated response of recipes
// @Description Paginated recipe response
type RecipePage struct {
	Recipes    []Recipe `json:"recipes"`
	Total      int64    `json:"total" example:"100"`
	Page       int      `json:"page" example:"1"`
	Limit      int      `json:"limit" example:"10"`
	TotalPages int      `json:"total_pages" example:"10"`
}
