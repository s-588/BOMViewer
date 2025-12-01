package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/internal/models"
	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) ProductPageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse initial filter parameters
	r.ParseForm()

	sort := r.FormValue("sort")

	// Get all products WITH materials
	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения списка продуктов", err)
		return
	}

	// Apply sorting
	sortConfig := helpers.ParseSortString(sort)
	helpers.SortProducts(products, sortConfig)

	// Get materials for filter dropdowns
	materials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		slog.Error("get all materials for filters", "error", err, "where", "ProductPageHandler")
	}

	// Parse material filters from form
	var filtersMaterials []int64
	if materialIDs := r.Form["materials"]; len(materialIDs) > 0 {
		filtersMaterials, _ = helpers.StringToInt64Slice(materialIDs)
	}

	templates.MainProductPage(products, templates.ProductTableArgs{
		Action:           "/products/table",
		Sort:             sort,
		FiltersMaterials: filtersMaterials,
		AllMaterials:     materials,
	}).Render(r.Context(), w)
}

func (h *Handler) ProductTableHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки формы", err)
		return
	}

	sort := r.FormValue("sort")

	// Get all products WITH materials
	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения списка продуктов", err)
		return
	}

	// Apply sorting
	sortConfig := helpers.ParseSortString(sort)
	helpers.SortProducts(products, sortConfig)

	// Get materials for filter dropdowns
	materials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		slog.Error("get all materials for filters", "error", err, "where", "ProductTableHandler")
		// Continue without materials
	}

	// Parse material filters from form
	var filtersMaterials []int64
	if materialIDs := r.Form["materials"]; len(materialIDs) > 0 {
		filtersMaterials, _ = helpers.StringToInt64Slice(materialIDs)
	}

	filteredProducts := helpers.FilterProducts(products, helpers.ProductFilterArgs{
		MaterialIDs: filtersMaterials,
	})

	templates.MainProductList(filteredProducts, templates.ProductTableArgs{
		Action:           "/products/table",
		Sort:             sort,
		FiltersMaterials: filtersMaterials,
		AllMaterials:     materials,
	}).Render(r.Context(), w)
}

func (h *Handler) ProductNewHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("name")
	if err := validateName(name); err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки имени продукта: "+err.Error(), err)
		return
	}

	description := r.FormValue("description")
	if err := validateDescription(description); err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки описания продукта: "+err.Error(), err)
		return
	}

	product := models.Product{
		Name:        name,
		Description: description,
	}

	productID, err := h.db.InsertProduct(r.Context(), product)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "внутренняя ошибка при создании продукта", err)
		return
	}

	materialIDs := r.Form["material_ids"]
	for _, idStr := range materialIDs {
		materialID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			helpers.WriteAndLogError(w, http.StatusInternalServerError, "внутренняя ошибка при обработке идентификатора материала", err)
			return
		}
		quantity := r.FormValue(fmt.Sprintf("quantity_%d", materialID))
		err = h.db.AddProductMaterial(r.Context(), productID, materialID, quantity)
		if err != nil {
			helpers.WriteAndLogError(w, http.StatusInternalServerError, "внутренняя ошибка при связывании продукта с материалом", err)
			return
		}
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func validateName(name string) error {
	if name == "" {
		return errors.New("имя обязательно для заполнения")
	}
	return nil
}

func validateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("описание слишком длинное")
	}
	return nil
}

func (h *Handler) ProductViewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}
	product, err := h.db.GetProductByID(r.Context(), id)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения продукта: "+err.Error(), err)
		return
	}
	files, err := h.db.GetProductFiles(r.Context(), id)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения файлов продукта: "+err.Error(), err)
		return
	}
	profilePicture, err := h.db.GetProductProfilePicture(r.Context(), id)
	if err != nil {
		// handle error
	}
	templates.ProductView(product, files, &profilePicture).Render(r.Context(), w)
}

func (h *Handler) ProductUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}

	// Parse form first!
	if err := r.ParseForm(); err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки формы", err)
		return
	}

	product, err := getProductFromRequest(r)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения продукта: "+err.Error(), err)
		return
	}

	// Update operations
	if product.Name != "" {
		err = h.db.UpdateProductName(r.Context(), id, product.Name)
		if err != nil {
			helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обновления имени продукта: "+err.Error(), err)
			return
		}
	}
	if product.Description != "" {
		err = h.db.UpdateProductDescription(r.Context(), id, product.Description)
		if err != nil {
			helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обновления описания продукта: "+err.Error(), err)
			return
		}
	}
	if len(product.Materials) != 0 {
		err = h.db.UpdateProductMaterials(r.Context(), id, product.Materials)
		if err != nil {
			helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обновления материалов продукта: "+err.Error(), err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/products/%d", id), http.StatusSeeOther)
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
	materialIDs := r.Form["material_ids"] // Changed from "material-ids"
	materialsList := make([]models.Material, 0, len(materialIDs))

	for _, idStr := range materialIDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		quantity := r.FormValue(fmt.Sprintf("quantity_%d", id)) // This matches your form
		materialsList = append(materialsList, models.Material{
			ID:       id,
			Quantity: quantity,
		})
	}
	return materialsList
}

func (h *Handler) ProductFilesListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}
	files, err := h.db.GetProductFiles(r.Context(), id)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения файлов продукта: "+err.Error(), err)
		return
	}
	templates.FileList(id, "product", files).Render(r.Context(), w)
}

func (h *Handler) ProductFileCreateHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("product-id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}
	fileID, err := strconv.ParseInt(r.URL.Query()["file-id"][0], 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора файла: "+err.Error(), err)
		return
	}
	h.db.InsertProductFile(r.Context(), productID, fileID)
}

func (h *Handler) ProductFileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.ParseInt(r.PathValue("file-id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора файла: "+err.Error(), err)
		return
	}
	h.db.DeleteFile(r.Context(), fileID)
}

func (h *Handler) ProductDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}
	err = h.db.DeleteProduct(r.Context(), id)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка удаления продукта: "+err.Error(), err)
		return
	}

	// For HTMX requests, return the updated table
	if r.Header.Get("HX-Request") == "true" {
		h.ProductTableHandler(w, r)
	} else {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}

func (h *Handler) ProductEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}

	product, err := h.db.GetProductByID(r.Context(), id)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения продукта: "+err.Error(), err)
		return
	}

	materials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения списка материалов: "+err.Error(), err)
		return
	}
	templates.ProductForm(product, materials, "edit").Render(r.Context(), w)
}

func (h *Handler) ProductCreateHandler(w http.ResponseWriter, r *http.Request) {
	materials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения списка материалов: "+err.Error(), err)
		return
	}
	templates.ProductForm(models.Product{}, materials, "new").Render(r.Context(), w)
}

func (h *Handler) ProductMaterialListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка обработки идентификатора продукта: "+err.Error(), err)
		return
	}
	materials, err := h.db.GetProductMaterials(r.Context(), id)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения материалов продукта: "+err.Error(), err)
		return
	}
	templates.MainMaterialPage(materials, templates.MaterialTableArgs{}).Render(r.Context(), w)
}

// In products.go - similar handler for products
func (h *Handler) SetProductProfilePicture(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	fileID, err := strconv.ParseInt(r.PathValue("fileID"), 10, 64)
	if err != nil {
		slog.Error("parse file id", "error", err)
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	err = h.db.UnsetProductProfilePicture(r.Context(), productID)
	if err != nil {
		slog.Warn("cannot unset product profile picture", "error", err, "where", "SetProductProfilePicture")
	}

	// THE PROBLEM: This tries to insert a duplicate association
	err = h.db.SetProductProfilePicture(r.Context(), productID, fileID)
	if err != nil {
		slog.Error("set product profile picture", "error", err)
		http.Error(w, "Error setting profile picture", http.StatusInternalServerError)
		return
	}

	h.ProductViewHandler(w, r)
}

func (h *Handler) RemoveProductProfilePicture(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.db.UnsetProductProfilePicture(r.Context(), productID)
	if err != nil {
		slog.Error("remove product profile picture", "error", err)
		http.Error(w, "Error removing profile picture", http.StatusInternalServerError)
		return
	}

	h.ProductViewHandler(w, r)
}
