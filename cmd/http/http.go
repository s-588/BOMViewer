package http

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/s-588/BOMViewer/cmd/config"
	"github.com/s-588/BOMViewer/cmd/http/handlers"
	"github.com/s-588/BOMViewer/cmd/http/middleware"
	"github.com/s-588/BOMViewer/internal/db"
)

var (
	//go:embed static/*
	embededStaticFiles embed.FS
)

type Server struct {
	ctx         context.Context
	mux         *http.ServeMux
	cancel      context.CancelFunc
	handler     *handlers.Handler
	cfg         *config.Config
	authManager *middleware.AuthManager
}

func NewServer(cancel context.CancelFunc, repo *db.Repository, cfg *config.Config) *Server {
	return &Server{
		cancel:      cancel,
		handler:     handlers.NewHandler(repo, cfg),
		mux:         http.NewServeMux(),
		cfg:         cfg,
		authManager: middleware.NewAuthManager(cfg.WebUIPassword),
	}
}

func (s *Server) Start(portChan chan int) error {
	s.setupPaths()

	port := fmt.Sprintf(":%d", s.cfg.ServerCfg.ServerPort)
	ls, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	portChan <- ls.Addr().(*net.TCPAddr).Port

	return http.Serve(ls, s.authManager.AuthMiddleware(s.mux))
}

func (s *Server) setupPaths() {
	slog.Info("setting up server endpoints")

	s.mux.HandleFunc("/", s.handler.RootPage)
	s.mux.Handle("/static/", http.FileServer(http.FS(embededStaticFiles)))
	s.mux.HandleFunc("/exit", s.stop)

	s.mux.HandleFunc("GET /search", s.handler.SearchHandler)                                       // return list of materials with checkboxes for forms
	s.mux.HandleFunc("GET /materials", s.handler.MaterialPageHandler)                              // Full page
	s.mux.HandleFunc("GET /materials/table", s.handler.MaterialTableHandler)                       // Table only (for HTMX)
	s.mux.HandleFunc("GET /materials/picker", s.handler.MaterialsPicker)                           // return list of materials with checkboxes for forms
	s.mux.HandleFunc("POST /materials", s.handler.MaterialNewHandler)                              // create new material, return new list of materials
	s.mux.HandleFunc("GET /materials/{id}", s.handler.MaterialViewHandler)                         // return material by id
	s.mux.HandleFunc("POST /materials/{id}", s.handler.MaterialUpdateHandler)                      // update material, return updated material
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

	s.mux.HandleFunc("GET /config", s.handler.ConfigPageHandler)
	s.mux.HandleFunc("POST /config", s.handler.UpdateConfigHandler)
	s.mux.HandleFunc("DELETE /config", s.handler.ResetConfigHandler)
	s.mux.HandleFunc("DELETE /config/{field}", s.handler.ResetConfigHandler)

	s.mux.HandleFunc("GET /login", s.handler.LoginPageHandler)
	s.mux.HandleFunc("POST /login", s.authManager.LoginHandler)
	s.mux.HandleFunc("DELETE /login", s.authManager.LogoutHandler)
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	s.cancel()
}
