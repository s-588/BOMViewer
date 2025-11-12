package handlers

import (
	"errors"
	"net/http"
	"log/slog"
	"strconv"

	"github.com/s-588/BOMViewer/internal/models"
	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) ProductListHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		slog.Error("get all products", "error", err, "where", "ProductListHandler")
		templates.InternalError("ошибка получения списка продуктов: " + err.Error()).Render(r.Context(), w)
		return
	}
	templates.ProductList(products).Render(r.Context(), w)
}

func (h *Handler) ProductNewHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if err := validateName(name); err != nil {
		slog.Error("Invalid product name", "name", name)
		templates.InternalError("ошибка обработки имени продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	materials := parseProductMaterials(r)
	product := models.Product{
		Name: name,
		Materials: materials,
		Description: r.FormValue("description"),
	}
	h.db.InsertProduct(r.Context(), product)
}

func validateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	return nil
}

func (h *Handler) ProductViewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductViewHandler")
		templates.InternalError("ошибка обработки идентификатора продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	product, err := h.db.GetProductByID(r.Context(), id)
	if err != nil {
		slog.Error("get product by id", "error", err, "where", "ProductViewHandler")
		templates.InternalError("ошибка получения продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	files, err := h.db.GetProductFiles(r.Context(), id)
	if err != nil {
		slog.Error("get product files", "error", err, "where", "ProductViewHandler")
		templates.InternalError("ошибка получения файлов продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	templates.ProductView(product, files).Render(r.Context(), w)
}

func (h *Handler) ProductUpdateHandler(w http.ResponseWriter, r *http.Request) {
	product, err := getProductFromRequest(r)
	if err != nil {
		slog.Error("get product from request", "error", err, "where", "ProductUpdateHandler")
		templates.InternalError("ошибка получения продукта: " + err.Error()).Render(r.Context(), w)
		return
	}

	if product.Name != "" {
		h.db.UpdateProductName(r.Context(), product.ID, product.Name)
	}
	if product.Description != "" {
		h.db.UpdateProductDescription(r.Context(), product.ID, product.Description)
	}
	if len(product.Materials) != 0 {
		h.db.UpdateProductMaterials(r.Context(), product.ID, product.Materials)
	}
}

func getProductFromRequest(r *http.Request) (models.Product, error) {
	var product models.Product
	name := r.FormValue("name")
	if err := validateName(name); err != nil {
		return product, err
	}
	product.Name = name
	product.Description = r.FormValue("description")
	materials := parseProductMaterials(r)
	product.Materials = materials
	return product, nil
}

func parseProductMaterials(r *http.Request) []models.Material {
	r.ParseForm()
	materialIDs := r.Form["material-ids"]
	materialCounts := r.Form["material-counts"]

	materialsList := make([]models.Material, 0)
	for i := 0; i < len(materialIDs); i++ {
		id, err := strconv.ParseInt(materialIDs[i],10,64)
		if err != nil{
			return materialsList
		}
		materialsList = append(materialsList, models.Material{
			ID: id,
			Quantity: materialCounts[i],
		})
	}
	return materialsList
}

func (h *Handler) ProductFilesListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductFilesListHandler")
		templates.InternalError("ошибка обработки идентификатора продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	files, err := h.db.GetProductFiles(r.Context(), id)
	if err != nil {
		slog.Error("get product files", "error", err, "where", "ProductFilesListHandler")
		templates.InternalError("ошибка получения файлов продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	templates.FileList(files).Render(r.Context(), w)
}

func (h *Handler) ProductFileCreateHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("product-id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductFileCreateHandler")
		templates.InternalError("ошибка обработки идентификатора продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	fileID, err := strconv.ParseInt(r.URL.Query()["file-id"][0], 10, 64)
	if err != nil {
		slog.Error("parse file id", "error", err, "where", "ProductFileCreateHandler")
		templates.InternalError("ошибка обработки идентификатора файла: " + err.Error()).Render(r.Context(), w)
		return
	}
	h.db.InsertProductFile(r.Context(), productID, fileID)
}

func (h *Handler) ProductFileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.ParseInt(r.PathValue("file-id"), 10, 64)
	if err != nil {
		slog.Error("parse file id", "error", err, "where", "ProductFileDeleteHandler")
		templates.InternalError("ошибка обработки идентификатора файла: " + err.Error()).Render(r.Context(), w)
		return
	}
	h.db.DeleteFile(r.Context(), fileID)
}

func (h *Handler) ProductDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductDeleteHandler")
		templates.InternalError("ошибка обработки идентификатора продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	err = h.db.DeleteProduct(r.Context(), id)
	if err != nil {
		slog.Error("delete product", "error", err, "where", "ProductDeleteHandler")
		templates.InternalError("ошибка удаления продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
	
func (h *Handler) ProductEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductEditHandler")
		templates.InternalError("ошибка обработки идентификатора продукта: " + err.Error()).Render(r.Context(), w)
		return
	}

	product, err := h.db.GetProductByID(r.Context(), id)
	if err != nil {
		slog.Error("get product by id", "error", err, "where", "ProductEditHandler")
		templates.InternalError("ошибка получения продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	templates.ProductForm(product, "edit").Render(r.Context(), w)
}

func (h *Handler) ProductCreateHandler(w http.ResponseWriter, r *http.Request) {
	templates.ProductForm(models.Product{}, "new").Render(r.Context(), w)
}

func (h *Handler) ProductMaterialListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductMaterialListHandler")
		templates.InternalError("ошибка обработки идентификатора продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	materials, err := h.db.GetProductMaterials(r.Context(), id)
	if err != nil {
		slog.Error("get product materials", "error", err, "where", "ProductMaterialListHandler")
		templates.InternalError("ошибка получения материалов продукта: " + err.Error()).Render(r.Context(), w)
		return
	}
	templates.MaterialList(materials).Render(r.Context(), w)
}
