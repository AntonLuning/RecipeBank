package pages

import (
	"fmt"
	"strconv"
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
