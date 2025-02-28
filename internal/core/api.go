package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/AntonLuning/RecipeBank/internal/core/service"
	"github.com/AntonLuning/RecipeBank/internal/core/storage"
	"github.com/AntonLuning/RecipeBank/pkg/core/models"
)

type apiFunc func(context.Context, http.ResponseWriter, *http.Request) error

type ApiServer struct {
	addr    string
	service service.Service
	mux     *http.ServeMux
}

func NewApiServer(addr string, service service.Service) *ApiServer {
	server := ApiServer{
		addr:    addr,
		service: service,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", server.v1Mux()))
	server.mux = mux

	return &server
}

func (s *ApiServer) Run() error {
	slog.Info("API server starting", "address", s.addr)

	return http.ListenAndServe(s.addr, s.mux)
}

func (s *ApiServer) v1Mux() http.Handler {
	v1Mux := http.NewServeMux()

	v1Mux.HandleFunc("GET /recipe", makeHTTPHandlerFunc(s.handleGetRecipes))
	v1Mux.HandleFunc("GET /recipe/{id}", makeHTTPHandlerFunc(s.handleGetRecipeByID))
	v1Mux.HandleFunc("POST /recipe", makeHTTPHandlerFunc(s.handlePostRecipe))
	v1Mux.HandleFunc("PUT /recipe/{id}", makeHTTPHandlerFunc(s.handlePutRecipe))
	v1Mux.HandleFunc("DELETE /recipe/{id}", makeHTTPHandlerFunc(s.handleDeleteRecipe))

	return v1Mux
}

func (s *ApiServer) handleGetRecipes(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var query models.GetRecipesQuery
	if err := s.parseQueryParams(r, &query); err != nil {
		return err
	}

	// Set defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	recipes, err := s.service.GetRecipes(ctx, query.Filter, query.Page, query.Limit)
	if err != nil {
		return err
	}

	return writeSuccessResponse(w, http.StatusOK, recipes)
}

func (s *ApiServer) handleGetRecipeByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

func (s *ApiServer) handlePostRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

func (s *ApiServer) handlePutRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

func (s *ApiServer) handleDeleteRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
			case errors.Is(err, ErrBadRequest):
				writeErrorResponse(w, http.StatusBadRequest, "bad_request", err.Error())
			case errors.Is(err, ErrInvalidQueryParams):
				writeErrorResponse(w, http.StatusBadRequest, "invalid_query_params", err.Error())
			case errors.Is(err, ErrMissingPathParam):
				writeErrorResponse(w, http.StatusBadRequest, "missing_path_param", err.Error())
			case errors.Is(err, ErrRequestBodyTooLarge):
				writeErrorResponse(w, http.StatusRequestEntityTooLarge, "request_too_large", err.Error())
			case errors.Is(err, service.ErrValidation):
				writeErrorResponse(w, http.StatusBadRequest, "validation_error", err.Error())
			case errors.Is(err, service.ErrInvalidInput):
				writeErrorResponse(w, http.StatusBadRequest, "invalid_input", err.Error())
			case errors.Is(err, storage.ErrInvalidID):
				writeErrorResponse(w, http.StatusBadRequest, "invalid_id", err.Error())
			case errors.Is(err, storage.ErrNotFound):
				writeErrorResponse(w, http.StatusNotFound, "not_found", err.Error())
			default:
				writeErrorResponse(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}
	}
}

// Helper functions
func (s *ApiServer) parseJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	return json.NewDecoder(r.Body).Decode(v)
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

func (s *ApiServer) parseQueryParams(r *http.Request, query *models.GetRecipesQuery) error {
	q := r.URL.Query()

	if pageStr := q.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return fmt.Errorf("%w: invalid page parameter: %v", ErrInvalidQueryParams, err)
		}
		query.Page = page
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return fmt.Errorf("%w: invalid limit parameter: %v", ErrInvalidQueryParams, err)
		}
		query.Limit = limit
	}

	// Parse filter parameters
	var filter models.RecipeFilter

	if title := q.Get("title"); title != "" {
		filter.Title = title
	}

	if cookTimeStr := q.Get("cook_time"); cookTimeStr != "" {
		cookTime, err := strconv.Atoi(cookTimeStr)
		if err != nil {
			return fmt.Errorf("%w: invalid cook_time parameter: %v", ErrInvalidQueryParams, err)
		}
		filter.CookTime = cookTime
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
