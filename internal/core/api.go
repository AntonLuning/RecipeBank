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

	"github.com/AntonLuning/RecipeBank/internal/core/middleware"
	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
	"github.com/AntonLuning/RecipeBank/pkg/models"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/AntonLuning/RecipeBank/docs" // Import generated swagger docs
)

type apiFunc func(http.ResponseWriter, *http.Request) error

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

	// API (Swagger) documentation routes
	mux.Handle("/docs/", httpSwagger.WrapHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})

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

	// AI-powered recipe creation
	v1Mux.HandleFunc("POST /recipe/ai/from-image", makeHTTPHandlerFunc(s.handlePostRecipeFromImage))
	v1Mux.HandleFunc("POST /recipe/ai/from-url", makeHTTPHandlerFunc(s.handlePostRecipeFromURL))

	return middleware.LoggingMiddleware(v1Mux)
}

// GetRecipes godoc
// @Summary Get all recipes
// @Description Get a paginated list of recipes with optional filtering
// @Tags recipes
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param title query string false "Filter by recipe title"
// @Param cook_time query int false "Filter by maximum cook time in minutes"
// @Param ingredients query string false "Filter by ingredient names (comma-separated)"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Success 200 {object} models.APIResponse{data=models.RecipePage} "Successful response"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid query parameters"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe [get]
func (s *APIServer) handleGetRecipes(w http.ResponseWriter, r *http.Request) error {
	var query models.GetRecipesQuery
	if err := s.parseQueryParams(r, &query); err != nil {
		return err
	}

	recipes, err := s.service.GetRecipes(r.Context(), query.Filter, query.Page, query.Limit)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, recipes)
}

// GetRecipeByID godoc
// @Summary Get recipe by ID
// @Description Get a specific recipe by its ID
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {object} models.APIResponse{data=models.Recipe} "Successful response"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid recipe ID"
// @Failure 404 {object} models.APIResponse{error=models.APIError} "Recipe not found"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe/{id} [get]
func (s *APIServer) handleGetRecipeByID(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("%w: id parameter is required", ErrMissingPathParam)
	}

	recipe, err := s.service.GetRecipe(r.Context(), id)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, recipe)
}

// PostRecipe godoc
// @Summary Create a new recipe
// @Description Create a new recipe with the provided information
// @Tags recipes
// @Accept json
// @Produce json
// @Param recipe body models.CreateRecipeRequest true "Recipe information"
// @Success 201 {object} models.APIResponse{data=models.Recipe} "Recipe created successfully"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid input data"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe [post]
func (s *APIServer) handlePostRecipe(w http.ResponseWriter, r *http.Request) error {
	var req models.CreateRecipeRequest
	if err := s.parseJSONBody(w, r, &req); err != nil {
		return err
	}

	recipe := createRecipeFromRequest(req)

	createdRecipe, err := s.service.CreateRecipe(r.Context(), recipe)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusCreated, createdRecipe)
}

// PutRecipe godoc
// @Summary Update a recipe
// @Description Update an existing recipe with the provided information
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Param recipe body models.UpdateRecipeRequest true "Updated recipe information"
// @Success 200 {object} models.APIResponse{data=models.Recipe} "Recipe updated successfully"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid input data or recipe ID"
// @Failure 404 {object} models.APIResponse{error=models.APIError} "Recipe not found"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe/{id} [put]
func (s *APIServer) handlePutRecipe(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("%w: id parameter is required", ErrMissingPathParam)
	}

	var req models.UpdateRecipeRequest
	if err := s.parseJSONBody(w, r, &req); err != nil {
		return err
	}

	recipe := createRecipeFromRequest(req)

	updatedRecipe, err := s.service.UpdateRecipe(r.Context(), id, recipe)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, updatedRecipe)
}

// DeleteRecipe godoc
// @Summary Delete a recipe
// @Description Delete a recipe by its ID
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 204 {object} models.APIResponse "Recipe deleted successfully"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid recipe ID"
// @Failure 404 {object} models.APIResponse{error=models.APIError} "Recipe not found"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe/{id} [delete]
func (s *APIServer) handleDeleteRecipe(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("%w: id parameter is required", ErrMissingPathParam)
	}

	if err := s.service.DeleteRecipe(r.Context(), id); err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusNoContent, nil)
}

// PostRecipeFromImage godoc
// @Summary Create recipe from image using AI
// @Description Create a new recipe by analyzing an image using AI
// @Tags ai-recipes
// @Accept json
// @Produce json
// @Param request body models.CreateRecipeFromImageRequest true "Image data and type"
// @Success 201 {object} models.APIResponse{data=models.Recipe} "Recipe created successfully from image"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid input data or AI processing error"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe/ai/from-image [post]
func (s *APIServer) handlePostRecipeFromImage(w http.ResponseWriter, r *http.Request) error {
	var req models.CreateRecipeFromImageRequest
	if err := s.parseJSONBody(w, r, &req); err != nil {
		return err
	}

	recipe, err := s.service.CreateRecipeFromImage(r.Context(), req.Image, req.ImageType)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusCreated, recipe)
}

// PostRecipeFromURL godoc
// @Summary Create recipe from URL using AI
// @Description Create a new recipe by analyzing content from a URL using AI
// @Tags ai-recipes
// @Accept json
// @Produce json
// @Param request body models.CreateRecipeFromUrlRequest true "URL to analyze"
// @Success 201 {object} models.APIResponse{data=models.Recipe} "Recipe created successfully from URL"
// @Failure 400 {object} models.APIResponse{error=models.APIError} "Invalid input data or AI processing error"
// @Failure 500 {object} models.APIResponse{error=models.APIError} "Internal server error"
// @Router /recipe/ai/from-url [post]
func (s *APIServer) handlePostRecipeFromURL(w http.ResponseWriter, r *http.Request) error {
	var req models.CreateRecipeFromUrlRequest
	if err := s.parseJSONBody(w, r, &req); err != nil {
		return err
	}

	recipe, err := s.service.CreateRecipeFromURL(r.Context(), req.URL)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusCreated, recipe)
}

func makeHTTPHandlerFunc(apiFn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := apiFn(w, r); err != nil {
			switch {
			case errors.Is(err, ErrJSONDecode):
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "invalid_json", "The request body contains invalid JSON", err)
			case errors.Is(err, ErrInvalidQueryParams):
				msg := "One or more query parameters are invalid"
				if paramErr := extractParamNameFromError(err.Error()); paramErr != "" {
					msg = fmt.Sprintf("Invalid query parameter: %s", paramErr)
				}
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "invalid_query_params", msg, err)
			case errors.Is(err, ErrMissingPathParam):
				msg := "A required path parameter is missing"
				if paramErr := extractParamNameFromError(err.Error()); paramErr != "" {
					msg = fmt.Sprintf("Missing required path parameter: %s", paramErr)
				}
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "missing_path_param", msg, err)
			case errors.Is(err, ErrRequestBodyTooLarge):
				writeErrorResponse(r.Context(), w, http.StatusRequestEntityTooLarge, "request_too_large", "The request body exceeds the maximum allowed size", err)
			case errors.Is(err, service.ErrValidation):
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "validation_error", extractValidationDetails(err.Error()), err)
			case errors.Is(err, service.ErrInvalidInput):
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "invalid_input", extractInputErrorDetails(err.Error()), err)
			case errors.Is(err, service.ErrAIUnsupported):
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "ai_unsupported", "AI processing is not supported/enabled", err)
			case errors.Is(err, service.ErrAI):
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "ai_error", "An error occurred while processing the AI request", err)
			case errors.Is(err, storage.ErrInvalidID):
				writeErrorResponse(r.Context(), w, http.StatusBadRequest, "invalid_id", "The provided ID is invalid or malformed", err)
			case errors.Is(err, storage.ErrNotFound):
				writeErrorResponse(r.Context(), w, http.StatusNotFound, "not_found", extractResourceTypeFromError(err.Error()), err)
			default:
				writeErrorResponse(r.Context(), w, http.StatusInternalServerError, "internal_error", "An internal server error occurred", err)
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
	return &models.Recipe{
		Title:       req.Title,
		Description: req.Description,
		Ingredients: req.Ingredients,
		Steps:       req.Steps,
		CookTime:    req.CookTime,
		Servings:    req.Servings,
		Tags:        req.Tags,
		Image:       req.Image,
	}
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

func writeErrorResponse(ctx context.Context, w http.ResponseWriter, status int, code string, message string, err error) error {
	slog.Error("Request failed", "request_id", middleware.GetRequestID(ctx), "error", err, "message", message)
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

func extractValidationDetails(errMsg string) string {
	return strings.TrimPrefix(errMsg, "validation error: ")
}

func extractInputErrorDetails(_ string) string {
	// TODO: This can be enhanced to parse specific input error types

	return "The provided input data is invalid or incomplete"
}

func extractResourceTypeFromError(errMsg string) string {
	resourceType := "resource" // Default resource type

	lowerMsg := strings.ToLower(errMsg)
	for _, knownType := range []string{"recipe", "ingredient", "tag"} {
		if strings.Contains(lowerMsg, knownType) {
			resourceType = knownType
			break
		}
	}

	return fmt.Sprintf("The requested %s was not found", resourceType)
}
