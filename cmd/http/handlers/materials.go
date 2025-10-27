package handlers

import (
	"net/http"
)

func (h *Handler) MaterialListHandler(w http.ResponseWriter, r *http.Request)        {}
func (h *Handler) MaterialNewHandler(w http.ResponseWriter, r *http.Request)         {}
func (h *Handler) MaterialViewHandler(w http.ResponseWriter, r *http.Request)        {}
func (h *Handler) MaterialUpdateHandler(w http.ResponseWriter, r *http.Request)      {}
func (h *Handler) MaterialFileListHandler(w http.ResponseWriter, r *http.Request)    {}
func (h *Handler) MaterialFileCreateHandler(w http.ResponseWriter, r *http.Request)  {}
func (h *Handler) MaterialFileDeleteHandler(w http.ResponseWriter, r *http.Request)  {}
func (h *Handler) MaterialDeleteHandler(w http.ResponseWriter, r *http.Request)      {}
func (h *Handler) MaterialEditHandler(w http.ResponseWriter, r *http.Request)        {}
func (h *Handler) MaterialCreateHandler(w http.ResponseWriter, r *http.Request)      {}
func (h *Handler) MaterialProductListHandler(w http.ResponseWriter, r *http.Request) {}
