package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Recipe struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Ingredients []Ingredient       `bson:"ingredients" json:"ingredients"`
	Steps       []string           `bson:"steps" json:"steps"`
	CookTime    int                `bson:"cook_time" json:"cook_time,omitempty"` // in minutes
	Servings    int                `bson:"servings" json:"servings,omitempty"`
	Tags        []string           `bson:"tags" json:"tags,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type Ingredient struct {
	Name     string `bson:"name" json:"name"`
	Quantity int    `bson:"quantity" json:"quantity"`
	Unit     string `bson:"unit,omitempty" json:"unit,omitempty"`
}

type RecipeFilter struct {
	Title           string   `json:"title,omitempty"`
	IngredientNames []string `json:"ingredient_names,omitempty"`
	CookTime        int      `json:"cook_time,omitempty"`
	Tags            []string `json:"tags,omitempty"`
}

type RecipePage struct {
	Recipes    []Recipe `json:"recipes"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}
