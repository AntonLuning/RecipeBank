package components

import (
	"fmt"
	"strconv"
)

func GetQueryValue(query map[string][]string, key string) string {
	if values, exists := query[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
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
