package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	initTestServer()
	defer server.Close()

	code := m.Run()

	os.Exit(code)
}

func initTestServer() {
	storage := storage.NewStorage() // TODO: Needs to be handled if require database

	recipeService := service.NewRecipeService(&storage)
	api := NewApiServer("", recipeService)

	server = httptest.NewServer(api.mux)
}

func TestCreateAndFetchRecipes(t *testing.T) {
	client := server.Client()

	_, err := execRequest(client, "GET", "/recipe", nil, 500)
	if err != nil {
		t.Fatalf("%v", err)
	}

	jsonData := []byte(`{
		"bad_data": "Test title 1"
	}`)
	_, err = execRequest(client, "POST", "/recipe", bytes.NewBuffer(jsonData), 500)
	if err != nil {
		t.Fatalf("%v", err)
	}

	jsonData = []byte(`{
		"title": "Test title 1"
	}`)
	body, err := execRequest(client, "POST", "/recipe", bytes.NewBuffer(jsonData), 200)
	if err != nil {
		t.Fatalf("%v", err)
	}
	id, ok := body["id"].(string)
	if !ok {
		t.Fatalf("reponse body does not include the key value pair: id")
	}

	_, err = execRequest(client, "GET", "/recipe/123", nil, 500)
	if err != nil {
		t.Fatalf("%v", err)
	}

	_, err = execRequest(client, "GET", fmt.Sprintf("/recipe/%s", id), nil, 200)
	if err != nil {
		t.Fatalf("%v", err)
	}

	jsonData = []byte(`{
		"title": "Test title 2"
	}`)
	_, err = execRequest(client, "POST", "/recipe", bytes.NewBuffer(jsonData), 200)
	if err != nil {
		t.Fatalf("%v", err)
	}

	body, err = execRequest(client, "GET", "/recipe", nil, 200)
	if err != nil {
		t.Fatalf("%v", err)
	}
	recipes, ok := body["recipes"].([]interface{})
	if !ok {
		t.Fatalf("reponse body does not include the key value pair: recipes")
	}
	if len(recipes) != 2 {
		t.Fatalf("number of recipes should be 2, got %v", len(recipes))
	}
}

func execRequest(client *http.Client, requestType string, path string, requestBody io.Reader, expectedStatusCode int) (map[string]interface{}, error) {
	rootPath := "/api/v1"

	var resp *http.Response
	var err error
	switch requestType {
	case "GET":
		resp, err = client.Get(server.URL + rootPath + path)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
	case "POST":
		resp, err = client.Post(server.URL+rootPath+path, "application/json", requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("Expected status %d, got %d", expectedStatusCode, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var actualBody map[string]interface{}
	if err := json.Unmarshal(body, &actualBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return actualBody, nil
}
