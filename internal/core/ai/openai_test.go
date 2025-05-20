package ai

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			wantModel: "gpt-4.1-mini-2025-04-14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.model)
			assert.NotNil(t, client)
			assert.Equal(t, tt.wantModel, client.model)
		})
	}
}

func TestAnalyzeImage(t *testing.T) {
	apiKey := getAPIKey(t)

	client := NewOpenAIClient(apiKey, "")

	// Get image path from environment variable or use default test image
	imagePath := os.Getenv("TEST_IMAGE_PATH")
	if imagePath == "" {
		t.Skip("Skipping test: TEST_IMAGE_PATH environment variable not set")
	}

	// Load image from file path
	imageFile, err := os.ReadFile(imagePath)
	require.NoError(t, err, "Failed to read test image file")

	imageBuffer := bytes.NewBuffer(imageFile)

	// Determine content type based on file extension
	contentType := ImageContentTypeJPEG
	if len(imagePath) > 4 && imagePath[len(imagePath)-4:] == ".png" {
		contentType = ImageContentTypePNG
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.AnalyzeImage(ctx, *imageBuffer, contentType, "What is this image? Keep it brief.")
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	t.Logf("Image analysis result: %s", result)
}

func TestAnalyzeURL(t *testing.T) {
	apiKey := getAPIKey(t)

	client := NewOpenAIClient(apiKey, "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.AnalyzeURL(ctx, "https://openai.com", "What is this website about? Keep it brief.")
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	t.Logf("URL analysis result: %s", result)
}

func TestAnalyzeImage_InvalidAPIKey(t *testing.T) {
	client := NewOpenAIClient("invalid-api-key", "")

	imageData := []byte("fake-image-data")
	imageBuffer := bytes.NewBuffer(imageData)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.AnalyzeImage(ctx, *imageBuffer, ImageContentTypeJPEG, "What is this image?")
	assert.Error(t, err)
}
