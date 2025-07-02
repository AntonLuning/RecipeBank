package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

type Renderable interface {
	Render(ctx context.Context, w io.Writer) error
}

func renderComponent(w http.ResponseWriter, r *http.Request, component Renderable) {
	var buf bytes.Buffer
	if err := component.Render(r.Context(), &buf); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

func fetchFromAPI[T any](apiBaseURL string, apiEndpoint string) (*T, error) {
	apiURL := fmt.Sprintf("%s/%s", apiBaseURL, strings.TrimPrefix(apiEndpoint, "/"))

	// Make GET request
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	// Check if the resource was not found
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("resource not found")
	}

	// Check if the response is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return convertResponseData[T](resp)
}

func createFromAPI[T any](apiBaseURL string, apiEndpoint string, jsonData []byte) (*T, error) {
	apiURL := fmt.Sprintf("%s/%s", apiBaseURL, strings.TrimPrefix(apiEndpoint, "/"))

	// Make POST request
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return convertResponseData[T](resp)
}

func updateFromAPI[T any](apiBaseURL string, apiEndpoint string, data *T) (*T, error) {
	apiURL := fmt.Sprintf("%s/%s", apiBaseURL, strings.TrimPrefix(apiEndpoint, "/"))

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create the PUT request
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return convertResponseData[T](resp)
}

func deleteFromAPI(apiBaseURL string, apiEndpoint string) error {
	apiURL := fmt.Sprintf("%s/%s", apiBaseURL, strings.TrimPrefix(apiEndpoint, "/"))

	// Create the DELETE request
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the resource was not found
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("resource not found")
	}

	// Check if the response is OK (200) or No Content (204)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

func convertResponseData[T any](resp *http.Response) (*T, error) {
	// Decode the response body to models.APIResponse
	var apiResponse models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Check if the API request was successful
	if !apiResponse.Success {
		errorMsg := "API request failed"
		if apiResponse.Error != nil {
			errorMsg = apiResponse.Error.Message
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// The API response data format should match map[string]interface{}
	responseData, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected API response format")
	}

	// Marshal the data to JSON and unmarshal it to the target struct
	// This is a workaround to convert the data to the correct type
	dataJSON, err := json.Marshal(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}
	var result T
	if err := json.Unmarshal(dataJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return &result, nil
}
