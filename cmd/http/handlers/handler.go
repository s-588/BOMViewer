package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/s-588/BOMViewer/internal/db"
	"github.com/s-588/BOMViewer/web/templates"
)

type Handler struct {
	db *db.Repository
}

func (h *Handler) RootPage(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(r.Context(), time.Second*10)
	err := templates.Index(r.Context()).Render(ctx, w)
	if err != nil {
		slog.Error("can't render welcome page", "error", err)
	}
}

func NewHandler(db *db.Repository)*Handler{
	return &Handler{
		db: db,
	 }
}