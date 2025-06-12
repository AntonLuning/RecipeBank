package pages

import (
	"fmt"
	"strconv"
)

// getQueryValue extracts a query parameter value
func getQueryValue(query map[string][]string, key string) string {
	if values, exists := query[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}

// getRecipeImageSrc returns the image source or a data URI fallback
func getRecipeImageSrc(image string) string {
	if image != "" {
		return image
	}
	// Return an inline SVG data URI as fallback to avoid network requests
	return "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjE5MiIgdmlld0JveD0iMCAwIDIwMCAxOTIiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSIyMDAiIGhlaWdodD0iMTkyIiBmaWxsPSIjRjNGNEY2Ii8+CjxwYXRoIGQ9Ik04NCA5NkM4NCA4Ny4xNjM0IDkxLjE2MzQgODAgMTAwIDgwQzEwOC44MzcgODAgMTE2IDg3LjE2MzQgMTE2IDk2QzExNiAxMDQuODM3IDEwOC44MzcgMTEyIDEwMCAxMTJDOTEuMTYzNCAxMTIgODQgMTA0LjgzNyA4NCA5NloiIGZpbGw9IiM5Q0EzQUYiLz4KPHBhdGggZD0iTTY4IDEyOEw4NC4zNDMxIDExMS42NTdDODcuNDY3MyAxMDguNTMzIDkyLjUzMjcgMTA4LjUzMyA5NS42NTY5IDExMS42NTdMMTMyIDEyOEg2OFoiIGZpbGw9IiM5Q0EzQUYiLz4KPC9zdmc+"
}

// formatCookTime formats cook time for display
func formatCookTime(cookTime int) string {
	return fmt.Sprintf("%d min", cookTime)
}

// formatServings formats servings for display
func formatServings(servings int) string {
	return fmt.Sprintf("%d servings", servings)
}

// formatExtraTags formats the extra tags count
func formatExtraTags(count int) string {
	return strconv.Itoa(count)
}

// formatPageNumber formats page number for display
func formatPageNumber(page int) string {
	return strconv.Itoa(page)
}

// shouldShowPage determines if a page number should be shown in pagination
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
