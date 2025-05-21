package ai

import (
	"context"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// OpenAIModel = "gpt-4.1-mini-2025-04-14"
	OpenAIModel = "gpt-4.1-2025-04-14"
)

func getAPIKey(t *testing.T) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY environment variable not set")
	}
	return apiKey
}

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		model     string
		wantModel string
	}{
		{
			name:      "with custom model",
			apiKey:    "test-api-key",
			model:     "custom-model",
			wantModel: "custom-model",
		},
		{
			name:      "with default model",
			apiKey:    "test-api-key",
			model:     "",
			wantModel: OpenAIModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAI(tt.apiKey, tt.model)
			assert.NotNil(t, client)
		})
	}
}

func TestAnalyzeImage(t *testing.T) {
	apiKey := getAPIKey(t)
	client := NewOpenAI(apiKey, OpenAIModel)

	// Get image path from environment variable or use default test image
	imagePath := os.Getenv("TEST_IMAGE_PATH")
	if imagePath == "" {
		t.Skip("Skipping test: TEST_IMAGE_PATH environment variable not set")
	}

	// Load image from file path
	imageFile, err := os.ReadFile(imagePath)
	require.NoError(t, err, "Failed to read test image file")

	base64Image := base64.StdEncoding.EncodeToString(imageFile)

	// Determine content type based on file extension
	contentType := ImageContentTypeJPEG
	if len(imagePath) > 4 && imagePath[len(imagePath)-4:] == ".png" {
		contentType = ImageContentTypePNG
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.AnalyzeRecipeImage(ctx, base64Image, contentType)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAnalyzeURL(t *testing.T) {
	apiKey := getAPIKey(t)
	client := NewOpenAI(apiKey, OpenAIModel)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.AnalyzeRecipeWebpage(ctx, "https://www.ica.se/recept/klassisk-lasagne-679675/")
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAnalyzeImage_InvalidAPIKey(t *testing.T) {
	client := NewOpenAI("invalid-api-key", OpenAIModel)

	imageData := []byte("fake-image-data")
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.AnalyzeRecipeImage(ctx, base64Image, ImageContentTypeJPEG)
	assert.Error(t, err)
}
