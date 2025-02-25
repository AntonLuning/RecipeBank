package models

type Recipe struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type RecipeFilter struct {
	Title string
}
