package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/s-588/BOMViewer/cmd/http/handlers"
	"github.com/s-588/BOMViewer/internal/db"
)

type Server struct {
	ctx     context.Context
	mux     *http.ServeMux
	cancel  context.CancelFunc
	handler *handlers.Handler
	Port    string
}

func NewServer(cancel context.CancelFunc, port string, repo *db.Repository) *Server {
	return &Server{
		cancel:  cancel,
		handler: handlers.NewHandler(repo),
		Port:    port,
		mux:     http.NewServeMux(),
	}
}

func (s *Server) Start() error {
	s.setupPaths()
	slog.Info("starting listening on port", "port", s.Port)
	return http.ListenAndServe(s.Port, s.mux)
}

func (s *Server) setupPaths() {
	slog.Info("setting up server endpoints")

	s.mux.HandleFunc("/", s.handler.RootPage)
	s.mux.Handle("/static/", http.FileServer(http.Dir("web/")))
	s.mux.HandleFunc("/exit", s.stop)

	s.mux.HandleFunc("GET /search", s.handler.SearchHandler)                                       // return list of materials with checkboxes for forms
	s.mux.HandleFunc("GET /materials", s.handler.MaterialPageHandler)                              // Full page
	s.mux.HandleFunc("GET /materials/table", s.handler.MaterialTableHandler)                       // Table only (for HTMX)
	s.mux.HandleFunc("GET /materials/picker", s.handler.MaterialsPicker)                           // return list of materials with checkboxes for forms
	s.mux.HandleFunc("POST /materials", s.handler.MaterialNewHandler)                              // create new material, return new list of materials
	s.mux.HandleFunc("GET /materials/{id}", s.handler.MaterialViewHandler)                         // return material by id
	s.mux.HandleFunc("POST /materials/{id}", s.handler.MaterialUpdateHandler)                      // update material, return updated material
	s.mux.HandleFunc("GET /materials/{id}/products", s.handler.MaterialProductListHandler)         // return list of products that use this material
	s.mux.HandleFunc("GET /materials/{id}/files", s.handler.MaterialFileListHandler)               // return list of pinned files
	s.mux.HandleFunc("POST /materials/{id}/upload-file", s.handler.MaterialFileUploadHandler)      // attach new file
	s.mux.HandleFunc("DELETE /materials/{id}/files/{fileID}", s.handler.MaterialFileDeleteHandler) // delete file
	s.mux.HandleFunc("DELETE /materials/{id}", s.handler.MaterialDeleteHandler)                    // delete material
	s.mux.HandleFunc("GET /materials/{id}/edit", s.handler.MaterialEditHandler)                    // return form for editing material
	s.mux.HandleFunc("GET /materials/new", s.handler.MaterialCreateHandler)                        // return form for creating new material

	s.mux.HandleFunc("GET /products", s.handler.ProductPageHandler)
	s.mux.HandleFunc("GET /products/table", s.handler.ProductTableHandler)
	s.mux.HandleFunc("POST /products", s.handler.ProductNewHandler)                              // create new product, return new list of products
	s.mux.HandleFunc("GET /products/{id}", s.handler.ProductViewHandler)                         // return product by id
	s.mux.HandleFunc("POST /products/{id}", s.handler.ProductUpdateHandler)                      // update product, return updated product
	s.mux.HandleFunc("GET /products/{id}/materials", s.handler.ProductMaterialListHandler)       // return list of materials that used in this product
	s.mux.HandleFunc("GET /products/{id}/files", s.handler.ProductFilesListHandler)              // return list of pinned files
	s.mux.HandleFunc("POST /products/{id}/upload-file", s.handler.ProductFileUploadHandler)      // attach new file
	s.mux.HandleFunc("DELETE /products/{id}/files/{fileID}", s.handler.ProductFileDeleteHandler) // delete file
	s.mux.HandleFunc("DELETE /products/{id}", s.handler.ProductDeleteHandler)                    // delete material
	s.mux.HandleFunc("GET /products/{id}/edit", s.handler.ProductEditHandler)                    // return form for editing product
	s.mux.HandleFunc("GET /products/new", s.handler.ProductCreateHandler)                        // return form for creating new product

	s.mux.HandleFunc("GET /files/{id}", s.handler.FileDownload)
	s.mux.HandleFunc("GET /files/preview/{id}", s.handler.FilePreview)
	s.mux.HandleFunc("POST /materials/{id}/set-profile-picture/{fileID}", s.handler.SetMaterialProfilePicture)
	s.mux.HandleFunc("POST /products/{id}/set-profile-picture/{fileID}", s.handler.SetProductProfilePicture)
	s.mux.HandleFunc("POST /materials/{id}/remove-profile-picture", s.handler.RemoveMaterialProfilePicture)
	s.mux.HandleFunc("POST /products/{id}/remove-profile-picture", s.handler.RemoveProductProfilePicture)

	s.mux.HandleFunc("GET /calculator", s.handler.CalculatorPageHandler)
	s.mux.HandleFunc("GET /calculator/products/{id}/materials", s.handler.CalculatorProductMaterialsHandler)
	s.mux.HandleFunc("POST /calculator/calculate", s.handler.CalculatorCalculateHandler)
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	s.cancel()
}
