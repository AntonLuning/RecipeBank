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
	detectedImageType, err := detectImageTypeFromBase64(image)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if detectedImageType != strings.ToLower(imageType) {
		if strings.ToLower(imageType) == "jpg" && detectedImageType != "jpeg" {
			return fmt.Errorf("image type mismatch: expected %s, got %s", imageType, detectedImageType)
		}
	}

	return nil
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

func detectImageTypeFromDataURI(imageData string) (string, error) {
	if imageData == "" {
		return "", nil // Empty image is valid (optional field)
	}

	// Check if it's a data URI format
	if strings.HasPrefix(imageData, "data:") {
		// Extract MIME type and base64 data from data URI
		parts := strings.Split(imageData, ",")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid data URI format")
		}

		// Extract MIME type from the first part (e.g., "data:image/jpeg;base64")
		headerPart := parts[0]
		if strings.Contains(headerPart, "image/jpeg") {
			// Validate by decoding the base64 part
			if _, err := base64.StdEncoding.DecodeString(parts[1]); err != nil {
				return "", fmt.Errorf("invalid base64 encoding: %w", err)
			}
			return "jpeg", nil
		} else if strings.Contains(headerPart, "image/png") {
			// Validate by decoding the base64 part
			if _, err := base64.StdEncoding.DecodeString(parts[1]); err != nil {
				return "", fmt.Errorf("invalid base64 encoding: %w", err)
			}
			return "png", nil
		} else {
			return "", fmt.Errorf("unsupported image type in data URI")
		}
	}

	return "", fmt.Errorf("invalid data URI format")
}
