package pages

import (
	"fmt"
	"strconv"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func getQueryValue(query map[string][]string, key string) string {
	if values, exists := query[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}

func getRecipeImageSrc(image string) string {
	if image != "" {
		return image
	}
	return "/assets/img/default_recipe_image.jpeg"
}

func formatCookTime(cookTime int) string {
	return fmt.Sprintf("%d min", cookTime)
}

func formatServings(servings int) string {
	return fmt.Sprintf("%d servings", servings)
}

func formatExtraTags(count int) string {
	return strconv.Itoa(count)
}

func formatPageNumber(page int) string {
	return strconv.Itoa(page)
}

func shouldShowPage(page, currentPage, totalPages int) bool {
	// Always show first and last page
	if page == 1 || page == totalPages {
		return true
	}

	// Show pages within 2 of current page
	if page >= currentPage-2 && page <= currentPage+2 {
		return true
	}

	return false
}

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
