package ui

import (
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/handlers"
)

func InitAssets(m *http.ServeMux, assetsPath string, isDebug bool) {
	fs := http.FileServer(http.Dir(assetsPath))
	m.Handle("GET /assets/", disableCacheInDevMode(http.StripPrefix("/assets/", fs), isDebug))
	m.Handle("GET /favicon.ico", serveFavicon(assetsPath))
}

func InitRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /", handlers.GetIndexPage)
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
