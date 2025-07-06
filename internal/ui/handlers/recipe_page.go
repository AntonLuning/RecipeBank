package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func GetRecipePage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract recipe ID from URL
		recipeID := r.PathValue("id")

		// Handle new recipe creation
		if recipeID == "new" || recipeID == "" {
			// Render recipe page with empty/new recipe
			component := pages.RecipePage(pages.RecipePageData{
				Recipe:   &models.Recipe{},
				EditMode: true,
			})
			renderComponent(w, r, component)
			return
		}

		// Check if edit mode is requested for existing recipes
		editMode := r.URL.Query().Get("edit") == "true"

		// Fetch recipe from API
		recipe, err := fetchRecipeFromAPI(apiURL, recipeID)
		if err != nil {
			// Handle error - render page with error state
			component := pages.RecipePage(pages.RecipePageData{Error: err.Error()})
			renderComponent(w, r, component)
			return
		}

		// Render page with recipe
		component := pages.RecipePage(pages.RecipePageData{
			Recipe:   recipe,
			EditMode: editMode,
		})
		renderComponent(w, r, component)
	}
}

func UpdateRecipe(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract recipe ID from URL
		recipeID := r.PathValue("id")
		if recipeID == "" {
			http.Error(w, "Recipe ID is required", http.StatusBadRequest)
			return
		}

		// Parse multipart form data (for file uploads)
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Get the current recipe first to merge changes
		currentRecipe, err := fetchRecipeFromAPI(apiURL, recipeID)
		if err != nil {
			http.Error(w, "Failed to fetch current recipe", http.StatusInternalServerError)
			return
		}

		// Update recipe from form
		if err := updateRecipeFromForm(currentRecipe, r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Send update to API
		recipe, err := updateRecipeViaAPI(apiURL, recipeID, currentRecipe)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update recipe: %v", err), http.StatusInternalServerError)
			return
		}

		// Render page with recipe
		component := pages.RecipePage(pages.RecipePageData{Recipe: recipe})
		renderComponent(w, r, component)
	}
}

func DeleteRecipe(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract recipe ID from URL
		recipeID := r.PathValue("id")
		if recipeID == "" {
			http.Error(w, "Recipe ID is required", http.StatusBadRequest)
			return
		}

		// Delete recipe via API
		err := deleteRecipeViaAPI(apiURL, recipeID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete recipe: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func fetchRecipeFromAPI(apiBaseURL string, recipeID string) (*models.Recipe, error) {
	apiEndpointWithID := fmt.Sprintf("/recipe/%s", recipeID)
	recipe, err := fetchFromAPI[models.Recipe](apiBaseURL, apiEndpointWithID)
	if err != nil {
		// Convert generic "resource not found" to more specific "recipe not found"
		if err.Error() == "resource not found" {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to fetch recipe: %w", err)
	}
	return recipe, nil
}

func updateRecipeViaAPI(apiBaseURL string, recipeID string, recipe *models.Recipe) (*models.Recipe, error) {
	apiEndpointWithID := fmt.Sprintf("/recipe/%s", recipeID)
	recipe, err := updateFromAPI(apiBaseURL, apiEndpointWithID, recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}
	return recipe, nil
}

func deleteRecipeViaAPI(apiBaseURL string, recipeID string) error {
	apiEndpointWithID := fmt.Sprintf("/recipe/%s", recipeID)
	err := deleteFromAPI(apiBaseURL, apiEndpointWithID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	return nil
}

func updateRecipeFromForm(recipe *models.Recipe, r *http.Request) error {
	recipe.Title = r.FormValue("title")
	recipe.Description = r.FormValue("description")

	cookTime, err := strconv.Atoi(r.FormValue("cookTime"))
	if err != nil {
		return fmt.Errorf("invalid cook time: %w", err)
	}
	recipe.CookTime = cookTime

	servings, err := strconv.Atoi(r.FormValue("servings"))
	if err != nil {
		return fmt.Errorf("invalid servings: %w", err)
	}
	recipe.Servings = servings

	if tags, exists := r.Form["tags"]; exists {
		recipe.Tags = tags
	}

	// Handle image - either from URL or file upload
	imageURL := r.FormValue("imageURL")
	if imageURL != "" {
		// User provided an image URL
		dataURI, err := downloadImage(imageURL)
		if err != nil {
			return fmt.Errorf("unable to download image: %w", err)
		}
		recipe.Image = dataURI
	} else {
		// Check for uploaded image file
		file, header, err := r.FormFile("imageFile")
		if err == nil {
			// File was uploaded
			defer file.Close()
			image, err := uploadImage(file, header)
			if err != nil {
				return fmt.Errorf("failed to upload image: %w", err)
			}
			recipe.Image = image
		}
		// If neither URL nor file provided, keep the existing image
	}

	// Parse ingredients from JSON
	ingredientsJSON := r.FormValue("ingredients")
	if ingredientsJSON != "" {
		var ingredients []models.Ingredient
		if err := json.Unmarshal([]byte(ingredientsJSON), &ingredients); err != nil {
			return fmt.Errorf("failed to parse ingredients: %w", err)
		}
		recipe.Ingredients = ingredients
	}

	if instructions, exists := r.Form["instructions"]; exists {
		recipe.Steps = instructions
	}

	return nil
}

func downloadImage(imageURL string) (string, error) {
	// Download image from URL
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Get content type from response header
	contentType := resp.Header.Get("Content-Type")

	// If content type is not available, try to detect from first few bytes
	if contentType == "" {
		if len(imageData) > 4 {
			// Check for common image formats
			switch {
			case len(imageData) >= 8 && string(imageData[:8]) == "\x89PNG\r\n\x1a\n":
				contentType = "image/png"
			case len(imageData) >= 3 && string(imageData[:3]) == "\xff\xd8\xff":
				contentType = "image/jpeg"
			case len(imageData) >= 6 && string(imageData[:6]) == "GIF87a" || string(imageData[:6]) == "GIF89a":
				contentType = "image/gif"
			case len(imageData) >= 12 && string(imageData[8:12]) == "WEBP":
				contentType = "image/webp"
			default:
				contentType = "image/jpeg" // Default fallback
			}
		} else {
			contentType = "image/jpeg" // Default fallback
		}
	}

	// Encode as data URI
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64Data), nil
}

func uploadImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Read the file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Determine image type from content type or filename
	imageType := "jpeg"
	if header.Header.Get("Content-Type") != "" {
		contentType := header.Header.Get("Content-Type")
		if strings.Contains(contentType, "png") {
			imageType = "png"
		} else if strings.Contains(contentType, "jpg") {
			imageType = "jpg"
		}
	} else if strings.HasSuffix(strings.ToLower(header.Filename), ".png") {
		imageType = "png"
	} else if strings.HasSuffix(strings.ToLower(header.Filename), ".jpg") {
		imageType = "jpg"
	}

	// Create data URI
	return fmt.Sprintf("data:image/%s;base64,%s", imageType, base64Image), nil
}
