package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func main() {
	// Check if image file path is provided as argument
	if len(os.Args) < 2 {
		fmt.Printf("Usage: go run %s <image_file_path>\n", os.Args[0])
		os.Exit(1)
	}

	imagePath := os.Args[1]

	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		os.Exit(1)
	}

	// Encode to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Create request payload
	req := models.CreateRecipeFromImageRequest{
		Image:     base64Image,
		ImageType: "jpeg",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Send POST request
	url := "http://localhost:9876/api/v1/recipe/ai/from-image" // TODO: use env variable or as argument
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Print response
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))
}
