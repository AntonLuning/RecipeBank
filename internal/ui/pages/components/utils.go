package components

import "fmt"

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
