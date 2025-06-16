package pages

import (
	"fmt"
	"strconv"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func getRecipeTitle(recipe *models.Recipe) string {
	if recipe != nil {
		return recipe.Title
	}
	return "Recipe Not Found"
}

func formatIngredient(ingredient models.Ingredient) string {
	if ingredient.Quantity > 0 && ingredient.Unit != "" {
		return fmt.Sprintf("%.1f %s %s", ingredient.Quantity, ingredient.Unit, ingredient.Name)
	} else if ingredient.Quantity > 0 {
		return fmt.Sprintf("%.1f %s", ingredient.Quantity, ingredient.Name)
	} else if ingredient.Unit != "" {
		return fmt.Sprintf("%s %s", ingredient.Unit, ingredient.Name)
	}
	return ingredient.Name
}

func formatStepNumber(step int) string {
	return strconv.Itoa(step)
}
