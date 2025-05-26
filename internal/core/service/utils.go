package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Create client with timeout to avoid hanging on slow responses
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create a HEAD request
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if status code indicates success (2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("URL could not be found or is not accessible")
	}

	return nil
}

func validateBase64Image(image string, imageType string) error {
	if image == "" {
		return fmt.Errorf("image data cannot be empty")
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	// Check if it's an image by looking at file signatures
	if len(data) < 4 {
		return fmt.Errorf("data too short to be a valid image")
	}

	// Check common image format signatures
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) && (strings.ToLower(imageType) == "jpeg" || strings.ToLower(imageType) == "jpg") {
		// JPEG/JPG signature
		return nil
	} else if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) && strings.ToLower(imageType) == "png" {
		// PNG signature
		return nil
	}

	return fmt.Errorf("unrecognized/unsupported image format")
}

func detectImageTypeFromBase64(image string) (string, error) {
	if image == "" {
		return "", nil // Empty image is valid (optional field)
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return "", fmt.Errorf("invalid base64 encoding: %w", err)
	}

	// Check if it's an image by looking at file signatures
	if len(data) < 4 {
		return "", fmt.Errorf("data too short to be a valid image")
	}

	// Check common image format signatures
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return "jpeg", nil
	} else if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		return "png", nil
	}

	return "", fmt.Errorf("unrecognized/unsupported image format (only JPEG and PNG are supported)")
}
