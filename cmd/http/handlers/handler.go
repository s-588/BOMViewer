package handlers

import (
	"log/slog"
	"net/http"

	"github.com/s-588/BOMViewer/internal/db"
	"github.com/s-588/BOMViewer/web/templates"
)

type Handler struct {
	db *db.Repository
}

func (h *Handler) RootPage(w http.ResponseWriter, r *http.Request) {
	materials, err := h.db.GetAllMaterial(r.Context())
	if err != nil {
		slog.Error("can't get materials", "error", err, "where", "RootPage")
		templates.InternalError("ошибка получения списка материалов").Render(r.Context(), w)
	}
	products, err := h.db.GetAllProducts(r.Context())
	if err != nil{
		slog.Error("can't get products", "error", err, "where", "RootPage")
		templates.InternalError("ошибка получения списка продуктов").Render(r.Context(), w)
	}
	err = templates.Index(r.Context(), materials, products).Render(r.Context(), w)
	if err != nil {
		slog.Error("can't render root page", "error", err, "where", "RootPage")
		templates.InternalError("ошибка отображения главной страницы").Render(r.Context(), w)
	}
}

func NewHandler(db *db.Repository) *Handler {
	return &Handler{
		db: db,
	}
}