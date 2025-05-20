package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type apiFunc func(context.Context, http.ResponseWriter, *http.Request) error

type APIServer struct {
	addr    string
	service service.Service
	mux     *http.ServeMux
}

func NewAPIServer(addr string, service service.Service) *APIServer {
	server := APIServer{
		addr:    addr,
		service: service,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", server.v1Mux()))
	server.mux = mux

	return &server
}

func (s *APIServer) Run() error {
	slog.Info("API server starting", "address", s.addr)

	return http.ListenAndServe(s.addr, s.mux)
}

func (s *APIServer) v1Mux() http.Handler {
	v1Mux := http.NewServeMux()

	v1Mux.HandleFunc("GET /recipe", makeHTTPHandlerFunc(s.handleGetRecipes))
	v1Mux.HandleFunc("GET /recipe/{id}", makeHTTPHandlerFunc(s.handleGetRecipeByID))
	v1Mux.HandleFunc("POST /recipe", makeHTTPHandlerFunc(s.handlePostRecipe))
	v1Mux.HandleFunc("PUT /recipe/{id}", makeHTTPHandlerFunc(s.handlePutRecipe))
	v1Mux.HandleFunc("DELETE /recipe/{id}", makeHTTPHandlerFunc(s.handleDeleteRecipe))

	return v1Mux
}

func (s *APIServer) handleGetRecipes(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var query models.GetRecipesQuery
	if err := s.parseQueryParams(r, &query); err != nil {
		return err
	}

	recipes, err := s.service.GetRecipes(ctx, query.Filter, query.Page, query.Limit)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, recipes)
}

func (s *APIServer) handleGetRecipeByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("%w: id parameter is required", ErrMissingPathParam)
	}

	recipe, err := s.service.GetRecipe(ctx, id)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, recipe)
}

func (s *APIServer) handlePostRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req models.CreateRecipeRequest
	if err := s.parseJSONBody(w, r, &req); err != nil {
		return err
	}

	recipe := createRecipeFromRequest(req)

	createdRecipe, err := s.service.CreateRecipe(ctx, recipe)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusCreated, createdRecipe)
}

func (s *APIServer) handlePutRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("%w: id parameter is required", ErrMissingPathParam)
	}

	var req models.UpdateRecipeRequest
	if err := s.parseJSONBody(w, r, &req); err != nil {
		return err
	}

	recipe := createRecipeFromRequest(req)

	updatedRecipe, err := s.service.UpdateRecipe(ctx, id, recipe)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, updatedRecipe)
}

func (s *APIServer) handleDeleteRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("%w: id parameter is required", ErrMissingPathParam)
	}

	if err := s.service.DeleteRecipe(ctx, id); err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusNoContent, nil)
}

func makeHTTPHandlerFunc(apiFn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		slog.Info("Incoming request", "method", r.Method, "path", r.URL.Path)

		if err := apiFn(ctx, w, r); err != nil {
			slog.Error("Request failed", "error", err)

			switch {
			case errors.Is(err, ErrJSONDecode):
				writeErrorResponse(w, http.StatusBadRequest, "invalid_json", "The request body contains invalid JSON")
			case errors.Is(err, ErrInvalidQueryParams):
				msg := "One or more query parameters are invalid"
				if paramErr := extractParamNameFromError(err.Error()); paramErr != "" {
					msg = fmt.Sprintf("Invalid query parameter: %s", paramErr)
				}
				writeErrorResponse(w, http.StatusBadRequest, "invalid_query_params", msg)
			case errors.Is(err, ErrMissingPathParam):
				msg := "A required path parameter is missing"
				if paramErr := extractParamNameFromError(err.Error()); paramErr != "" {
					msg = fmt.Sprintf("Missing required path parameter: %s", paramErr)
				}
				writeErrorResponse(w, http.StatusBadRequest, "missing_path_param", msg)
			case errors.Is(err, ErrRequestBodyTooLarge):
				writeErrorResponse(w, http.StatusRequestEntityTooLarge, "request_too_large", "The request body exceeds the maximum allowed size")
			case errors.Is(err, service.ErrValidation):
				writeErrorResponse(w, http.StatusBadRequest, "validation_error", extractValidationDetails(err.Error()))
			case errors.Is(err, service.ErrInvalidInput):
				writeErrorResponse(w, http.StatusBadRequest, "invalid_input", extractInputErrorDetails(err.Error()))
			case errors.Is(err, storage.ErrInvalidID):
				writeErrorResponse(w, http.StatusBadRequest, "invalid_id", "The provided ID is invalid or malformed")
			case errors.Is(err, storage.ErrNotFound):
				writeErrorResponse(w, http.StatusNotFound, "not_found", fmt.Sprintf(
					"The requested %s was not found", extractResourceTypeFromError(err.Error()),
				))
			default:
				writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "An internal server error occurred")
			}
		}
	}
}

// Helper functions
func (s *APIServer) parseJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJSONDecode, err)
	}
	return nil
}

func createRecipeFromRequest(req models.RecipeRequest) *models.Recipe {
	recipe := &models.Recipe{
		Title:       req.Title,
		Description: req.Description,
		Ingredients: req.Ingredients,
		Steps:       req.Steps,
		CookTime:    req.CookTime,
		Servings:    req.Servings,
		Tags:        req.Tags,
	}
	return recipe
}

func writeJSON(w http.ResponseWriter, statusCode int, content any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(content)
}

func writeSuccessResponse(w http.ResponseWriter, status int, data interface{}) error {
	return writeJSON(w, status, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

func writeErrorResponse(w http.ResponseWriter, status int, code, message string) error {
	return writeJSON(w, status, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

func (s *APIServer) parseQueryParams(r *http.Request, query *models.GetRecipesQuery) error {
	q := r.URL.Query()

	// Parse pagination parameters with helper function
	var err error
	query.Page, err = parseIntParam(q, "page", 1)
	if err != nil {
		return err
	}

	query.Limit, err = parseIntParam(q, "limit", 10)
	if err != nil {
		return err
	}

	// Parse filter parameters
	var filter models.RecipeFilter

	filter.Title = q.Get("title")

	filter.CookTime, err = parseIntParam(q, "cook_time", 0)
	if err != nil {
		return err
	}

	if ingredients := q.Get("ingredients"); ingredients != "" {
		filter.IngredientNames = strings.Split(ingredients, ",")
	}

	if tags := q.Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	query.Filter = filter

	return nil
}

func parseIntParam(q url.Values, key string, defaultValue int) (int, error) {
	str := q.Get(key)
	if str == "" {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("%w: %s parameter is invalid", ErrInvalidQueryParams, key)
	}

	return val, nil
}

// Helper functions for extracting user-safe error details

// extractParamNameFromError extracts the name of the parameter from the error message
func extractParamNameFromError(errMsg string) string {
	if strings.Contains(errMsg, "parameter") {
		parts := strings.Split(errMsg, "parameter")
		if len(parts) > 1 {
			paramPart := strings.TrimSpace(parts[1])
			paramParts := strings.Split(paramPart, " ")
			if len(paramParts) > 0 {
				return paramParts[len(paramParts)-1]
			}
		}
	}
	return ""
}

// extractValidationDetails creates a user-friendly validation error message
func extractValidationDetails(errMsg string) string {
	return strings.TrimPrefix(errMsg, "validation error: ")
}

// extractInputErrorDetails creates a user-friendly input error message
func extractInputErrorDetails(_ string) string {
	// This can be enhanced to parse specific input error types

	return "The provided input data is invalid or incomplete"
}

// extractResourceTypeFromError attempts to extract the resource type from not found errors
func extractResourceTypeFromError(errMsg string) string {
	// Default resource type
	resourceType := "resource"

	lowerMsg := strings.ToLower(errMsg)
	for _, knownType := range []string{"recipe", "ingredient", "tag"} {
		if strings.Contains(lowerMsg, knownType) {
			resourceType = knownType
			break
		}
	}

	return resourceType
}
