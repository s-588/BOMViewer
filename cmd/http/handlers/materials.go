package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/internal/models"
	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) MaterialPageHandler(w http.ResponseWriter, r *http.Request) {
	units, err := h.db.GetAllUnits(r.Context())
	if err != nil {
		slog.Error("can't get units", "error", err, "where", "MaterialListHandler")
		templates.InternalError("ошибка получения списка единиц измерения для фильтров").Render(r.Context(), w)
		return
	}

	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		slog.Error("can't get products", "error", err, "where", "MaterialListHandler")
		templates.InternalError("ошибка получения списка продуктов для фильтров").Render(r.Context(), w)
		return
	}

	// Get initial materials (unfiltered)
	materials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		slog.Error("can't get materials", "error", err, "where", "MaterialListHandler")
		templates.InternalError("ошибка получения списка материалов").Render(r.Context(), w)
		return
	}

	filteredMaterials := helpers.FilterMaterials(materials, helpers.MaterialFilterArgs{})
	sortConfig := helpers.ParseSortString("name")
	helpers.SortMaterials(filteredMaterials, sortConfig)

	templates.MainMaterialPage(filteredMaterials, templates.MaterialTableArgs{
		Action:      "/materials/table",
		Sort:        "name",
		AllUnits:    units,
		AllProducts: products,
		Selected:    make(map[int64]bool),
		Quantities:  make(map[int64]string),
	}).Render(r.Context(), w)
}

func (h *Handler) MaterialTableHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data (includes both URL query and form parameters)
	if err := r.ParseForm(); err != nil {
		// handle error
	}

	// Get all parameters from form data
	sort := r.FormValue("sort")
	primaryOnly := r.FormValue("primary_only") == "1"

	var filtersUnits, filtersProducts []int64
	if units := r.Form["units"]; len(units) > 0 {
		filtersUnits, _ = helpers.StringToInt64Slice(units)
	}
	if products := r.Form["products"]; len(products) > 0 {
		filtersProducts, _ = helpers.StringToInt64Slice(products)
	}

	// Get all materials and apply filters
	allMaterials, err := h.db.GetAllMaterials(r.Context())
	if err != nil {
		slog.Error("can't get materials", "error", err, "where", "MaterialTableHandler")
		templates.InternalError("ошибка получения списка материалов").Render(r.Context(), w)
		return
	}

	// Apply filtering and sorting using your helper functions
	filteredMaterials := helpers.FilterMaterials(allMaterials, helpers.MaterialFilterArgs{
		PrimaryOnly: primaryOnly,
		ProductIDs:  filtersProducts,
		UnitIDs:     filtersUnits,
	})

	// Apply sorting
	sortConfig := helpers.ParseSortString(sort)
	helpers.SortMaterials(filteredMaterials, sortConfig)

	allUnits, err := h.db.GetAllUnits(r.Context())
	if err != nil {
		slog.Error("can't get units", "error", err, "where", "MaterialTableHandler")
		templates.InternalError("ошибка получения списка единиц измерения для фильтров").Render(r.Context(), w)
		return
	}

	allProducts, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		slog.Error("can't get products", "error", err, "where", "MaterialTableHandler")
		templates.InternalError("ошибка получения списка продуктов для фильтров").Render(r.Context(), w)
		return
	}

	args := templates.MaterialTableArgs{
		Action:          "/materials/table",
		Sort:            sort,
		FiltersUnits:    filtersUnits,
		FiltersProducts: filtersProducts,
		PrimaryOnly:     primaryOnly,
		AllUnits:        allUnits,
		AllProducts:     allProducts,
		Quantities:      make(map[int64]string),
		Selected:        make(map[int64]bool),
	}

	slog.Debug("materia table args", "args", args)

	slog.Debug("materila table filtered materials", "materials", filteredMaterials)

	// Return only the table, not the full page
	templates.MainMaterialList(filteredMaterials, args).Render(r.Context(), w)
}

func (h *Handler) MaterialNewHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	primaryName := r.FormValue("primary_name")
	names := r.Form["other_names"]
	if err := validateNames(names); err != nil {
		slog.Error("validate names", "error", err, "where", "MaterialNewHandler")
	}
	if primaryName == "" && len(names) == 0 {
		templates.InternalError("хотя бы одно имя должно быть указано").Render(r.Context(), w)
	}
	if primaryName != "" && !slices.Contains(names, primaryName) {
		names = append(names, primaryName)
	}

	unitID, err := strconv.ParseInt(r.FormValue("unit_id"), 10, 64)
	if err != nil {
		slog.Error("unit error", "error", err, "where", "MaterialNewHandler")
		templates.NotFoundError("внутреняя ошибка").Render(r.Context(), w)
	}
	unit, err := h.db.GetUnitByID(r.Context(), unitID)
	if err != nil {
		slog.Error("unit error", "error", err, "where", "MaterialNewHandler")
		templates.NotFoundError("единица измерения не существует").Render(r.Context(), w)
	}

	material := models.Material{
		Names:       names,
		PrimaryName: primaryName,
		Description: r.FormValue("description"),
		Unit:        unit,
	}
	material, err = h.db.InsertMaterial(r.Context(), material)
	if err != nil {
		slog.Error("can't insert material", "error", err, "where", "MaterialNewHandler")
		templates.InternalError("внутренняя ошибка").Render(r.Context(), w)
	}

	productsIds := r.Form["product_ids"]
	for _, id := range productsIds {
		productID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			slog.Error("parse product id", "error", err, "where", "MaterialNewHandler")
			templates.InternalError("ошибка обработки идентификатора продукта: "+err.Error()).Render(r.Context(), w)
			return
		}
		quantity := r.FormValue(fmt.Sprintf("quantity_%d", productID))
		err = h.db.AddProductMaterial(r.Context(), productID, material.ID, quantity)
		if err != nil {
			slog.Error("can't add product to material", "error", err, "where", "MaterialNewHandler")
			templates.InternalError("внутренняя ошибка").Render(r.Context(), w)
			return
		}
	}

	http.Redirect(w, r, "/materials", http.StatusSeeOther)
}

func parseUnit(unit string) (models.Unit, error) {
	if unit == "" {
		return models.Unit{}, errors.New("единица измерения обязательна")
	}
	return models.Unit{
		Name: unit,
	}, nil
}

func validateNames(names []string) error {
	if len(names) < 1 {
		return errors.New("хотябы одно имя должно быть")
	}
	for _, name := range names {
		if len(name) < 2 || name == "" || len(name) > 200 {
			return errors.New("длина имени должна быть от 2 до 200 символов")
		}
	}
	return nil
}

func (h *Handler) MaterialViewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// handle error
	}
	material, err := h.db.GetMaterialByID(r.Context(), id)
	if err != nil {
		// handle error
	}
	files, err := h.db.GetMaterialFiles(r.Context(), id)
	if err != nil {
		// handle error
	}
	profilePicture, err := h.db.GetMaterialProfilePicture(r.Context(), id)
	if err != nil {
		// handle error
	}

	templates.MaterialView(material, files, &profilePicture).Render(r.Context(), w)
}

func (h *Handler) MaterialUpdateHandler(w http.ResponseWriter, r *http.Request) {
	material, err := getMaterialFromRequest(r)
	if err != nil {
		slog.Error("get material from request", "error", err, "where", "MaterialUpdateHandler")
		templates.InternalError("ошибка получения материала: "+err.Error()).Render(r.Context(), w)
		return
	}

	if material.Description != "" {
		h.db.UpdateMaterialDescription(r.Context(), material.ID, material.Description)
	}
	if material.Unit.Name != "" {
		h.db.UpdateMaterialUnit(r.Context(), material.ID, material.Unit.Name)
	}
	if len(material.Names) != 0 {
		h.db.UpdateMaterialNames(r.Context(), material.ID, material.PrimaryName, material.Names)
	}
}

func getMaterialFromRequest(r *http.Request) (models.Material, error) {
	var material models.Material
	r.ParseForm()
	primaryName := r.FormValue("primary-name")
	names := r.Form["names"]
	if err := validateNames(names); err != nil {
		return material, err
	}
	material.PrimaryName = primaryName
	material.Names = names
	material.Description = r.FormValue("description")
	var err error
	material.Unit, err = parseUnit(r.FormValue("unit"))
	if err != nil {
		return material, err
	}
	return material, nil
}

func (h *Handler) MaterialFileListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err, "where", "MaterialFileListHandler")
		templates.InternalError("ошибка обработки идентификатора материала: "+err.Error()).Render(r.Context(), w)
		return
	}
	files, err := h.db.GetMaterialFiles(r.Context(), id)
	if err != nil {
		slog.Error("get material files", "error", err, "where", "MaterialFileListHandler")
		templates.InternalError("ошибка получения файлов материала: "+err.Error()).Render(r.Context(), w)
		return
	}
	templates.FileList(id, "material", files).Render(r.Context(), w)
}

func (h *Handler) MaterialFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("MaterialFileUploadHandler called", "method", r.Method, "path", r.URL.Path)

	materialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err, "where", "MaterialFileUploadHandler")
		templates.InternalError("ошибка обработки идентификатора материала").Render(r.Context(), w)
		return
	}
	slog.Info("materialID", "id", materialID)

	// Handle file upload - now accepts any file type
	uploadedFile, err := h.fileUpload.HandleFileUpload(r, "file")
	if err != nil {
		slog.Error("file upload error", "error", err, "where", "MaterialFileUploadHandler")
		templates.InternalError("ошибка загрузки файла: "+err.Error()).Render(r.Context(), w)
		return
	}
	slog.Info("file uploaded", "name", uploadedFile.Name, "path", uploadedFile.Path, "mime", uploadedFile.MimeType)

	// Determine file type based on MIME type
	fileType := "document"
	if strings.HasPrefix(uploadedFile.MimeType, "image/") {
		fileType = "image"
	}
	slog.Info("file type determined", "type", fileType)

	// Save file record to database
	fileID, err := h.db.InsertFile(r.Context(), models.File{
		Name:     uploadedFile.Name,
		Path:     uploadedFile.Path,
		MimeType: uploadedFile.MimeType,
		FileType: fileType,
	})
	if err != nil {
		// Clean up uploaded file
		h.fileUpload.DeleteFile(uploadedFile.Path)
		slog.Error("insert file error", "error", err, "where", "MaterialFileUploadHandler")
		templates.InternalError("ошибка сохранения файла в базе данных").Render(r.Context(), w)
		return
	}
	slog.Info("file saved to database", "fileID", fileID)

	// Link file to material
	err = h.db.InsertMaterialFile(r.Context(), materialID, fileID)
	if err != nil {
		// Clean up uploaded file and database record
		h.fileUpload.DeleteFile(uploadedFile.Path)
		h.db.DeleteFile(r.Context(), fileID)
		slog.Error("link file to material error", "error", err, "where", "MaterialFileUploadHandler")
		templates.InternalError("ошибка привязки файла к материалу").Render(r.Context(), w)
		return
	}
	slog.Info("file linked to material", "materialID", materialID, "fileID", fileID)

	// Get updated file list and return the files section
	files, err := h.db.GetMaterialFiles(r.Context(), materialID)
	if err != nil {
		slog.Error("get material files error", "error", err, "where", "MaterialFileUploadHandler")
		templates.InternalError("ошибка получения списка файлов").Render(r.Context(), w)
		return
	}
	slog.Info("retrieved material files", "count", len(files))

	// Return ONLY the FileList component
	slog.Info("rendering FileList template")
	templates.FileList(materialID, "materials", files).Render(r.Context(), w)
	slog.Info("FileList template rendered successfully")
}

func (h *Handler) MaterialFileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	materialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err, "where", "MaterialFileDeleteHandler")
		templates.InternalError("ошибка обработки идентификатора материала").Render(r.Context(), w)
		return
	}

	fileID, err := strconv.ParseInt(r.PathValue("fileID"), 10, 64)
	if err != nil {
		slog.Error("parse file id", "error", err, "where", "MaterialFileDeleteHandler")
		templates.InternalError("ошибка обработки идентификатора файла").Render(r.Context(), w)
		return
	}

	// Get file info before deletion
	file, err := h.db.GetFileByID(r.Context(), fileID)
	if err != nil {
		slog.Error("get file error", "error", err, "where", "MaterialFileDeleteHandler")
		// Continue with deletion anyway
	} else {
		// Delete physical file if we found it
		if err := h.fileUpload.DeleteFile(file.Path); err != nil {
			slog.Error("delete physical file error", "error", err, "where", "MaterialFileDeleteHandler")
		}
	}

	// Delete from database - this should handle both files_materials and files table due to CASCADE
	err = h.db.DeleteFile(r.Context(), fileID)
	if err != nil {
		slog.Error("delete file error", "error", err, "where", "MaterialFileDeleteHandler")
		templates.InternalError("ошибка удаления файла").Render(r.Context(), w)
		return
	}

	// Return updated file list - IMPORTANT: Return FileList template, not the full page
	files, err := h.db.GetMaterialFiles(r.Context(), materialID)
	if err != nil {
		slog.Error("get material files error", "error", err, "where", "MaterialFileDeleteHandler")
		templates.InternalError("ошибка получения списка файлов").Render(r.Context(), w)
		return
	}

	templates.FileList(materialID, "materials", files).Render(r.Context(), w)
}

func (h *Handler) MaterialDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err, "where", "MaterialDeleteHandler")
		templates.InternalError("ошибка обработки идентификатора материала: "+err.Error()).Render(r.Context(), w)
		return
	}
	err = h.db.DeleteMaterial(r.Context(), id)
	if err != nil {
		slog.Error("delete material", "error", err, "where", "MaterialDeleteHandler")
		templates.InternalError("ошибка удаления материала: "+err.Error()).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/materials", http.StatusSeeOther)
}

func (h *Handler) MaterialEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err, "where", "MaterialEditHandler")
		templates.InternalError("ошибка обработки идентификатора материала: "+err.Error()).Render(r.Context(), w)
		return
	}

	material, err := h.db.GetMaterialByID(r.Context(), id)
	if err != nil {
		slog.Error("get material by id", "error", err, "where", "MaterialEditHandler")
		templates.InternalError("ошибка получения материала: "+err.Error()).Render(r.Context(), w)
		return
	}

	units, err := h.db.GetAllUnits(r.Context())
	if err != nil {
		slog.Error("get all units", "error", err, "where", "MaterialEditHandler")
		templates.InternalError("ошибка получения списка единиц измерения: "+err.Error()).Render(r.Context(), w)
		return
	}

	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		slog.Error("get all products", "error", err, "where", "MaterialEditHandler")
		templates.InternalError("ошибка получения списка продуктов: "+err.Error()).Render(r.Context(), w)
		return
	}
	templates.MaterialForm(material, units, products, "edit").Render(r.Context(), w)
}

func (h *Handler) MaterialCreateHandler(w http.ResponseWriter, r *http.Request) {
	units, err := h.db.GetAllUnits(r.Context())
	if err != nil {
		slog.Error("get all units", "error", err, "where", "MaterialCreateHandler")
		templates.InternalError("ошибка получения списка единиц измерения: "+err.Error()).Render(r.Context(), w)
		return
	}
	products, err := h.db.GetAllProducts(r.Context())
	if err != nil {
		slog.Error("get all products", "error", err, "where", "MaterialCreateHandler")
		templates.InternalError("ошибка получения списка продуктов: "+err.Error()).Render(r.Context(), w)
		return
	}
	templates.MaterialForm(models.Material{}, units, products, "new").Render(r.Context(), w)
}

func (h *Handler) MaterialProductListHandler(w http.ResponseWriter, r *http.Request) {
	//TODO:
	// id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	// if err != nil {
	// 	slog.Error("parse material id", "error", err, "where", "MaterialProductListHandler")
	// 	templates.InternalError("ошибка обработки идентификатора материала: "+err.Error()).Render(r.Context(), w)
	// 	return
	// }
	// products, err := h.db.GetMaterialProducts(r.Context(), id)
	// if err != nil {
	// 	slog.Error("get material products", "error", err, "where", "MaterialProductListHandler")
	// 	templates.InternalError("ошибка получения списка продуктов материала: "+err.Error()).Render(r.Context(), w)
	// 	return
	// }
	// templates.ProductList(products).Render(r.Context(), w)
}

func (h *Handler) MaterialsPicker(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	materials, _, err := h.db.SearchAll(r.Context(), q, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows := make([]templates.ProductMaterialRow, 0, len(materials))
	for _, material := range materials {
		rows = append(rows, templates.ProductMaterialRow{
			Material: material,
		})
	}
	templates.MaterialTableForProduct(rows, templates.MaterialTableArgs{}).Render(r.Context(), w)
}

// In materials.go
func (h *Handler) SetMaterialProfilePicture(w http.ResponseWriter, r *http.Request) {
	materialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err)
		http.Error(w, "Invalid material ID", http.StatusBadRequest)
		return
	}

	fileID, err := strconv.ParseInt(r.PathValue("fileID"), 10, 64)
	if err != nil {
		slog.Error("parse file id", "error", err)
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	err = h.db.UnsetMaterialProfilePicture(r.Context(), materialID)
	if err != nil {
		slog.Warn("cannot unset material profile picture", "error", err, "where", "SetMaterialProfilePicture")
	}

	err = h.db.SetMaterialProfilePicture(r.Context(), materialID, fileID)
	if err != nil {
		slog.Error("set material profile picture", "error", err)
		http.Error(w, "Error setting profile picture", http.StatusInternalServerError)
		return
	}

	h.MaterialViewHandler(w, r)
}

func (h *Handler) RemoveMaterialProfilePicture(w http.ResponseWriter, r *http.Request) {
	materialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err)
		http.Error(w, "Invalid material ID", http.StatusBadRequest)
		return
	}

	err = h.db.UnsetMaterialProfilePicture(r.Context(), materialID)
	if err != nil {
		slog.Error("set material profile picture", "error", err)
		http.Error(w, "Error removing profile picture", http.StatusInternalServerError)
		return
	}

	h.MaterialViewHandler(w, r)
}
