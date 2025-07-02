package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func CreateRecipeFromImage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get the image file from form
		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Image file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content
		imageData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read image file", http.StatusInternalServerError)
			return
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

		// Create request payload
		requestPayload := models.CreateRecipeFromImageRequest{
			Image:     base64Image,
			ImageType: imageType,
		}

		// Call API to create recipe from image
		recipe, err := createRecipeFromImageAPI(apiURL, requestPayload)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create recipe from image: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to the recipe detail page in edit mode
		http.Redirect(w, r, fmt.Sprintf("/recipe/%s?edit=true", recipe.ID.Hex()), http.StatusSeeOther)
	}
}

func createRecipeFromImageAPI(apiBaseURL string, request models.CreateRecipeFromImageRequest) (*models.Recipe, error) {
	apiEndpoint := "/recipe/ai/from-image"

	// Prepare request data
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	recipe, err := createFromAPI[models.Recipe](apiBaseURL, apiEndpoint, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe from image: %w", err)
	}

	return recipe, nil
}
