package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) MaterialImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	materialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "Неверный идентификатор материала", err)
		return
	}

	// Handle file upload
	uploadedFile, err := h.fileUpload.HandleFileUpload(r, "image")
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка загрузки файла: "+err.Error(), err)
		return
	}

	// Save file record to database
	fileID, err := h.db.InsertFile(r.Context(), *uploadedFile)
	if err != nil {
		// Clean up uploaded file
		h.fileUpload.DeleteFile(uploadedFile.Path)
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка сохранения файла в базе данных", err)
		return
	}

	// Link file to material
	err = h.db.InsertMaterialFile(r.Context(), materialID, fileID)
	if err != nil {
		// Clean up uploaded file and database record
		h.fileUpload.DeleteFile(uploadedFile.Path)
		h.db.DeleteFile(r.Context(), fileID)
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка привязки файла к материалу", err)
		return
	}

	h.MaterialViewHandler(w, r)
}

func (h *Handler) ProductFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "Неверный идентификатор изделия", err)
		return
	}

	// Handle file upload - using the same fileUpload service as materials
	uploadedFile, err := h.fileUpload.HandleFileUpload(r, "file") // or "image" if you want to restrict to images
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка загрузки файла: "+err.Error(), err)
		return
	}

	// Save file record to database
	fileID, err := h.db.InsertFile(r.Context(), *uploadedFile)
	if err != nil {
		// Clean up uploaded file
		h.fileUpload.DeleteFile(uploadedFile.Path)
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка сохранения файла в базе данных", err)
		return
	}

	// Link file to product
	err = h.db.InsertProductFile(r.Context(), productID, fileID)
	if err != nil {
		// Clean up uploaded file and database record
		h.fileUpload.DeleteFile(uploadedFile.Path)
		h.db.DeleteFile(r.Context(), fileID)
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка привязки файла к изделию", err)
		return
	}

	// Return updated product view
	h.ProductViewHandler(w, r)
}

func (h *Handler) ProductImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "Неверный идентификатор продукта", err)
		return
	}

	// Handle file upload (same logic as material)
	uploadedFile, err := h.fileUpload.HandleFileUpload(r, "image")
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка загрузки файла: "+err.Error(), err)
		return
	}

	// Save file record to database
	fileID, err := h.db.InsertFile(r.Context(), *uploadedFile)
	if err != nil {
		h.fileUpload.DeleteFile(uploadedFile.Path)
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка сохранения файла в базе данных", err)
		return
	}

	// Link file to product
	err = h.db.InsertProductFile(r.Context(), fileID, productID)
	if err != nil {
		err := h.fileUpload.DeleteFile(uploadedFile.Path)
		if err != nil {
			slog.Error("cannot delete file from filesystem", "error", err, "where", "ProductImageUploadHandler")
		}
		err = h.db.DeleteFile(r.Context(), fileID)
		if err != nil {
			slog.Error("cannot delete file from database", "error", err, "where", "ProductImageUploadHandler")
		}
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка привязки файла к продукту", err)
		return
	}

	// Get updated file list
	files, err := h.db.GetProductFiles(r.Context(), productID)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "ошибка получения списка файлов", err)
		return
	}

	err = templates.FileList(productID, "product", files).Render(r.Context(), w)
	if err != nil {
		slog.Error("can't render file list", "error", err, "where", "ProductImageUploadHandler")
	}
}

func (h *Handler) FilePreview(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "Неверный идентификатор файла", err)
		return
	}

	file, err := h.db.GetFileByID(r.Context(), fileID)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusNotFound, "Файл не найден", err)
		return
	}

	// If it's an image, serve it directly
	if file.FileType == "image" {
		http.ServeFile(w, r, file.Path)
		return
	}

	// For non-image files, you might want to show a generic preview
	// For now, we'll serve a generic file icon or redirect to download
	http.Redirect(w, r, "/files/"+strconv.FormatInt(fileID, 10), http.StatusSeeOther)
}

func (h *Handler) FileDownload(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "Неверный идентификатор файла", err)
		return
	}

	file, err := h.db.GetFileByID(r.Context(), fileID)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusNotFound, "Файл не найден", err)
		return
	}

	// Set appropriate headers for download
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.Path)
}
