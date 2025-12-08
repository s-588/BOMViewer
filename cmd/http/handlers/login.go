package handlers

import (
	"net/http"

	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.LoginPage().Render(r.Context(), w)
}
