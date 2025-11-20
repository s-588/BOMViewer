package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/s-588/BOMViewer/internal/db"
	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	// view := r.URL.Query().Get("view")
	dataType := r.URL.Query().Get("type")
	sort := r.URL.Query().Get("sort")
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 10
	}
	unitsFilter := "[" + strings.Join(r.URL.Query()["units"], ",") + "]"
	json.NewEncoder(w).Encode(unitsFilter)

	switch dataType {
	case "materials":
		_, err := h.db.SearchMaterials(r.Context(), db.SearchParams{
			Query:  query,
			Sort:   sort,
			Limit:  limit,
			Filter: json.RawMessage(unitsFilter),
		})
		if err != nil {
			slog.Error("can't search materials", "error", err, "where", "SearchHandler")
			templates.InternalError("ошибка поиска материалов").Render(r.Context(), w)
		}
	case "products":
		_, err := h.db.SearchProducts(r.Context(), db.SearchParams{
			Query: query,
			Sort:  sort,
			Limit: limit,
		})
		if err != nil {
			slog.Error("can't search products", "error", err, "where", "SearchHandler")
			templates.InternalError("ошибка поиска продуктов").Render(r.Context(), w)
		}
	default:
		materials, products, err := h.db.SearchAll(r.Context(), db.SearchParams{
			Query: query,
			Sort:  sort,
			Limit: limit,
		})
		if err != nil {
			slog.Error("can't search all", "error", err, "where", "SearchHandler")
			templates.InternalError("ошибка поиска всего").Render(r.Context(), w)
		}

		templates.SearchResults("all", "main", materials, products)
	}
}
