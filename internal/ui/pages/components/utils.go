package components

import (
	"fmt"
	"net/url"
	"strconv"
)

func GetQueryValue(query map[string][]string, key string) string {
	if values, exists := query[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}

func BuildPaginationQueryString(page int, searchQuery map[string][]string) string {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))

	// Add existing search parameters
	if title := GetQueryValue(searchQuery, "title"); title != "" {
		params.Set("title", title)
	}
	if ingredientNames := GetQueryValue(searchQuery, "ingredient_names"); ingredientNames != "" {
		params.Set("ingredient_names", ingredientNames)
	}
	if cookTime := GetQueryValue(searchQuery, "cook_time"); cookTime != "" {
		params.Set("cook_time", cookTime)
	}
	if tags := GetQueryValue(searchQuery, "tags"); tags != "" {
		params.Set("tags", tags)
	}

	return params.Encode()
}

func GetRecipeImageSrc(image string) string {
	if image != "" {
		return image
	}
	return "/assets/img/default_recipe_image.jpeg"
}

func FormatCookTime(cookTime int) string {
	return fmt.Sprintf("%d min", cookTime)
}

func FormatServings(servings int) string {
	return fmt.Sprintf("%d servings", servings)
}

func FormatExtraTags(count int) string {
	return strconv.Itoa(count)
}

func FormatPageNumber(page int) string {
	return strconv.Itoa(page)
}

func ShouldShowPage(page, currentPage, totalPages int) bool {
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
