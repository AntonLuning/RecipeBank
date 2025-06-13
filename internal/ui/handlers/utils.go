package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AntonLuning/RecipeBank/pkg/models"
)

func fetchFromAPI[T any](apiBaseURL string, apiEndpoint string) (*T, error) {
	apiURL := fmt.Sprintf("%s/%s", apiBaseURL, strings.TrimPrefix(apiEndpoint, "/"))
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("resource not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResponse models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if !apiResponse.Success {
		errorMsg := "API request failed"
		if apiResponse.Error != nil {
			errorMsg = apiResponse.Error.Message
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	// The API response data should match the expected type T
	responseData, ok := apiResponse.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected API response format")
	}

	// Convert the map back to the target struct
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
