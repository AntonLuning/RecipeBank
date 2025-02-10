package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/core/service"
)

type apiFunc func(context.Context, http.ResponseWriter, *http.Request) error

type ApiServer struct {
	addr    string
	service service.Service
}

func NewApiServer(addr string, service service.Service) *ApiServer {
	return &ApiServer{
		addr:    addr,
		service: service,
	}
}

func (s *ApiServer) Run() error {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", s.v1Mux()))

	slog.Info("API server starting", "address", s.addr)

	return http.ListenAndServe(s.addr, mux)
}

func (s *ApiServer) v1Mux() http.Handler {
	v1Mux := http.NewServeMux()

	v1Mux.HandleFunc("GET /recipe", makeHTTPHandlerFunc(s.handleGetRecipe))
	v1Mux.HandleFunc("GET /recipe/{id}", makeHTTPHandlerFunc(s.handleGetRecipeByID))
	v1Mux.HandleFunc("POST /recipe", makeHTTPHandlerFunc(s.handlePostRecipe))
	v1Mux.HandleFunc("PUT /recipe", makeHTTPHandlerFunc(s.handlePutRecipe))

	return v1Mux
}

func (s *ApiServer) handleGetRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	return writeJSON(w, http.StatusOK, "Hello Recipe")
}

func (s *ApiServer) handleGetRecipeByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	recipe, err := s.service.GetRecipe(ctx, id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, recipe.Title)
}

func (s *ApiServer) handlePostRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, "Hello POST Recipe")
}

func (s *ApiServer) handlePutRecipe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, "Hello PUT Recipe")
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
