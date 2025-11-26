package handlers

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/s-588/BOMViewer/internal/db"
	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/web/templates"
)

type Handler struct {
	db         *db.Repository
	fileUpload *helpers.FileUploadConfig
}

func (h *Handler) RootPage(w http.ResponseWriter, r *http.Request) {
	// Get materials for the main page (you might want to limit this or show featured items)
	materials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		slog.Error("can't get materials", "error", err, "where", "RootPage")
		templates.InternalError("ошибка получения списка материалов").Render(r.Context(), w)
		return
	}
	sortCfg := helpers.ParseSortString("name")
	helpers.SortMaterials(materials, sortCfg)

	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		slog.Error("can't get products", "error", err, "where", "RootPage")
		templates.InternalError("ошибка получения списка продуктов").Render(r.Context(), w)
		return
	}

	// Get units and products for the table controls
	units, err := h.db.GetAllUnits(r.Context())
	if err != nil {
		slog.Error("can't get units", "error", err, "where", "RootPage")
		templates.InternalError("ошибка получения единиц измерения").Render(r.Context(), w)
		return
	}

	// Create proper TableControlsArgs
	tableArgs := templates.MaterialTableArgs{
		Action:      "/materials/table",
		Sort:        "name",
		AllUnits:    units,
		AllProducts: products,
	}

	err = templates.Index(r.Context(), materials, tableArgs).Render(r.Context(), w)
	if err != nil {
		slog.Error("can't render root page", "error", err, "where", "RootPage")
		templates.InternalError("ошибка отображения главной страницы").Render(r.Context(), w)
	}
}

func NewHandler(db *db.Repository) *Handler {
	os.MkdirAll("uploads", 0755)
	return &Handler{
		db:         db,
		fileUpload: helpers.NewFileUploadConfig("uploads"),
	}
}
