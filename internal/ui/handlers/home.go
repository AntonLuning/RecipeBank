package handlers

import (
	"net/http"

	"github.com/AntonLuning/RecipeBank/internal/ui/pages"
)

func GetHomePage(apiURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		component := pages.HomePage()
		renderComponent(w, r, component)
	}
}
