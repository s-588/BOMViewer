package helpers

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/s-588/BOMViewer/internal/models"
)

var AllowedTypes = []string{
            "image/jpeg", "image/png", "image/gif", "image/webp",

            "application/pdf",
            "application/msword", // .doc
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document", // .docx
            "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // .xlsx
            "application/vnd.openxmlformats-officedocument.presentationml.presentation", // .pptx

            "text/plain",
            "application/rtf",

            "application/zip",
            "application/x-zip-compressed",
}

type FileUploadConfig struct {
	UploadDir     string
	MaxUploadSize int64
}

func NewFileUploadConfig(uploadDir string) *FileUploadConfig {
	return &FileUploadConfig{
		UploadDir:     uploadDir,
		MaxUploadSize: 100 * 1024 * 1024, // 100MB
	}
}

func (c *FileUploadConfig) HandleFileUpload(r *http.Request, formFieldName string) (*models.File, error) {
	// Parse multipart form
	if err := r.ParseMultipartForm(c.MaxUploadSize); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Get file from form
	file, header, err := r.FormFile(formFieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from form: %w", err)
	}
	defer file.Close()

	// Check file size
	if header.Size > c.MaxUploadSize {
		return nil, fmt.Errorf("file too large: %d KiB, max allowed: %d KiB", header.Size/1024, c.MaxUploadSize/1024)
	}

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer)
	if !c.isAllowedType(mimeType) {
		return nil, fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// Reset file reader
	if _, err = file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(c.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Try to get extension from MIME type
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[0]
		} else {
			ext = ".bin"
		}
	}

	uniqueName := generateUniqueFileName() + ext
	filePath := filepath.Join(c.UploadDir, uniqueName)

	// Create the file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Write the file
	if _, err = io.Copy(dst, file); err != nil {
		// Clean up if write fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Determine file type
	fileType := "document"
	if strings.HasPrefix(mimeType, "image/") {
		fileType = "image"
	}

	return &models.File{
		Name:     header.Filename,
		Path:     filePath,
		MimeType: mimeType,
		FileType: fileType,
	}, nil
}

func (c *FileUploadConfig) isAllowedType(mimeType string) bool {
	return slices.Contains(AllowedTypes, mimeType)
}

func generateUniqueFileName() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (c *FileUploadConfig) DeleteFile(filePath string) error {
	return os.Remove(filePath)
}
