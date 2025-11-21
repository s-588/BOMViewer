package handlers

import (
	"github.com/s-588/BOMViewer/web/templates"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := cleanFTS5Query(r.URL.Query().Get("q"))
	// view := r.URL.Query().Get("view")
	dataType := r.URL.Query().Get("type")
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 10
	}
	if len(q) == 0 {
		return
	}
	slog.Debug("search query:", "query", q)

	switch dataType {
	case "materials":
		_, err := h.db.SearchMaterials(r.Context(), q, limit)
		if err != nil {
			slog.Error("can't search materials", "error", err, "where", "SearchHandler")
			templates.InternalError("ошибка поиска материалов").Render(r.Context(), w)
		}
	case "products":
		_, err := h.db.SearchProducts(r.Context(), q, limit)
		if err != nil {
			slog.Error("can't search products", "error", err, "where", "SearchHandler")
			templates.InternalError("ошибка поиска продуктов").Render(r.Context(), w)
		}
	default:
		materials, products, err := h.db.SearchAll(r.Context(), q, limit)
		if err != nil {
			slog.Error("can't search all", "error", err, "where", "SearchHandler")
			templates.InternalError("ошибка поиска всего").Render(r.Context(), w)
		}

		slog.Debug("search results:", "materials", materials, "products", products)
		err = templates.SearchResults("all", "main", materials, products).Render(r.Context(), w)
		if err != nil {
			slog.Error("can't render search results", "error", err)
		}
	}
}

// Helper function to clean FTS5 query
func cleanFTS5Query(query string) string {
	if query == "" {
		return ""
	}

	specialChars := `"'\,.?!@#$%^&*()-+={}[]|:;<>`
	var result strings.Builder

	for _, char := range query {
		if strings.ContainsRune(specialChars, char) {
			continue
		}
		result.WriteRune(char)
	}

	return result.String()
}
