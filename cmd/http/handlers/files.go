package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) MaterialImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	materialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse material id", "error", err, "where", "MaterialImageUploadHandler")
		templates.InternalError("ошибка обработки идентификатора материала").Render(r.Context(), w)
		return
	}

	// Handle file upload
	uploadedFile, err := h.fileUpload.HandleFileUpload(r, "image")
	if err != nil {
		slog.Error("file upload error", "error", err, "where", "MaterialImageUploadHandler")
		templates.InternalError("ошибка загрузки файла: "+err.Error()).Render(r.Context(), w)
		return
	}

	// Save file record to database
	fileID, err := h.db.InsertFile(r.Context(), *uploadedFile)
	if err != nil {
		// Clean up uploaded file
		h.fileUpload.DeleteFile(uploadedFile.Path)
		slog.Error("insert file error", "error", err, "where", "MaterialImageUploadHandler")
		templates.InternalError("ошибка сохранения файла в базе данных").Render(r.Context(), w)
		return
	}

	// Link file to material
	err = h.db.InsertMaterialFile(r.Context(), materialID, fileID)
	if err != nil {
		// Clean up uploaded file and database record
		h.fileUpload.DeleteFile(uploadedFile.Path)
		h.db.DeleteFile(r.Context(), fileID)
		slog.Error("link file to material error", "error", err, "where", "MaterialImageUploadHandler")
		templates.InternalError("ошибка привязки файла к материалу").Render(r.Context(), w)
		return
	}

	// Get updated file list and return the files section
	files, err := h.db.GetMaterialFiles(r.Context(), materialID)
	if err != nil {
		slog.Error("get material files error", "error", err, "where", "MaterialImageUploadHandler")
		templates.InternalError("ошибка получения списка файлов").Render(r.Context(), w)
		return
	}

	err = templates.FileList(materialID, "material", files).Render(r.Context(), w)
	if err != nil {
		slog.Error("can't render file list", "error", err, "where", "MaterialImageUploadHandler")
	}
}

func (h *Handler) ProductImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		slog.Error("parse product id", "error", err, "where", "ProductImageUploadHandler")
		templates.InternalError("ошибка обработки идентификатора продукта").Render(r.Context(), w)
		return
	}

	// Handle file upload (same logic as material)
	uploadedFile, err := h.fileUpload.HandleFileUpload(r, "image")
	if err != nil {
		slog.Error("file upload error", "error", err, "where", "ProductImageUploadHandler")
		templates.InternalError("ошибка загрузки файла: "+err.Error()).Render(r.Context(), w)
		return
	}

	// Save file record to database
	fileID, err := h.db.InsertFile(r.Context(), *uploadedFile)
	if err != nil {
		h.fileUpload.DeleteFile(uploadedFile.Path)
		slog.Error("insert file error", "error", err, "where", "ProductImageUploadHandler")
		templates.InternalError("ошибка сохранения файла в базе данных").Render(r.Context(), w)
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
		slog.Error("link file to product error", "error", err, "where", "ProductImageUploadHandler")
		templates.InternalError("ошибка привязки файла к продукту").Render(r.Context(), w)
		return
	}

	// Get updated file list
	files, err := h.db.GetProductFiles(r.Context(), productID)
	if err != nil {
		slog.Error("get product files error", "error", err, "where", "ProductImageUploadHandler")
		templates.InternalError("ошибка получения списка файлов").Render(r.Context(), w)
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
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := h.db.GetFileByID(r.Context(), fileID)
	if err != nil {
		slog.Error("get file error", "error", err, "where", "FilePreview")
		http.Error(w, "File not found", http.StatusNotFound)
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
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := h.db.GetFileByID(r.Context(), fileID)
	if err != nil {
		slog.Error("get file error", "error", err, "where", "FileDownload")
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set appropriate headers for download
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file.Path)
}
