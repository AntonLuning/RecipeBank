package ui

import (
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/handlers"
)

func InitAssets(m *http.ServeMux, assetsPath string, isDebug bool) {
	fs := http.FileServer(http.Dir(assetsPath))
	m.Handle("GET /favicon.ico", serveFavicon(assetsPath))
	m.Handle("GET /assets/", disableCacheInDevMode(http.StripPrefix("/assets/", fs), isDebug))
}

func InitRoutes(m *http.ServeMux, apiURL string) {
	m.HandleFunc("GET /", handlers.GetHomePage(apiURL))                                       // Home page
	m.HandleFunc("GET /partial/recipes-overview", handlers.GetRecipesOverviewPartial(apiURL)) // Recipes overview partial
	m.HandleFunc("GET /partial/recipes", handlers.GetRecipesGridPartial(apiURL))              // Recipes grid partial

	m.HandleFunc("GET /recipe/{id}", handlers.GetRecipePage(apiURL))   // Recipe detail page
	m.HandleFunc("POST /recipe/{id}", handlers.UpdateRecipe(apiURL))   // Update recipe endpoint
	m.HandleFunc("DELETE /recipe/{id}", handlers.DeleteRecipe(apiURL)) // Delete recipe endpoint

	// m.HandleFunc("POST /recipe", handlers.CreateRecipe(apiURL))                     // Create recipe endpoint
	m.HandleFunc("POST /recipe/from-url", handlers.CreateRecipeFromURL(apiURL))     // Create recipe from URL endpoint
	m.HandleFunc("POST /recipe/from-image", handlers.CreateRecipeFromImage(apiURL)) // Create recipe from image endpoint
}

func serveFavicon(assetsPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, assetsPath+"/img/favicon.ico")
	})
}

func disableCacheInDevMode(h http.Handler, isDebug bool) http.Handler {
	if !isDebug {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}
