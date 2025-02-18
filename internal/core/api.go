package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/core/service"
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

	v1Mux.HandleFunc("GET /recipe", makeHTTPHandlerFunc(s.handleGetRecipe))
	v1Mux.HandleFunc("GET /recipe/{id}", makeHTTPHandlerFunc(s.handleGetRecipeByID))
	v1Mux.HandleFunc("POST /recipe", makeHTTPHandlerFunc(s.handlePostRecipe))
	// v1Mux.HandleFunc("PUT /recipe/{id}", makeHTTPHandlerFunc(s.handlePutRecipe))
	// v1Mux.HandleFunc("DELETE /recipe/{id}", makeHTTPHandlerFunc(s.handlePutRecipe))

	return v1Mux
}

func (s *ApiServer) handleGetRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	filter := "" // TODO: include an option to send a filter in the request

	recipes, err := s.service.GetRecipes(ctx, filter)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{"recipes": recipes})
}

func (s *ApiServer) handleGetRecipeByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("id argument is empty or missing")
	}

	recipe, err := s.service.GetRecipe(ctx, id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, recipe)
}

func (s *ApiServer) handlePostRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Limit the size of the request body to prevent potential abuse
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit[5]

	var data models.PostRecipeData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return err
	}

	id, err := s.service.CreateRecipe(ctx, data)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, models.PostRecipeResponse{
		ID: id,
	})
}

func makeHTTPHandlerFunc(apiFn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		slog.Info("Incoming request", "method", r.Method, "path", r.URL.Path)

		if err := apiFn(ctx, w, r); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal server error"})
		}
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, content any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(content)
}
