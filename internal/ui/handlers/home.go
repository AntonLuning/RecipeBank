package handlers

import (
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
)

func GetHomePage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		component := pages.HomePage()
		if err := component.Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
	}
}
