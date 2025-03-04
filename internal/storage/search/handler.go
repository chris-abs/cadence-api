package search

import (
	"encoding/json"
	"net/http"

	"github.com/chrisabs/storage/internal/middleware"
	"github.com/chrisabs/storage/internal/models"
	"github.com/gorilla/mux"
)

type Handler struct {
    service        *Service
    authMiddleware *middleware.AuthMiddleware
}

func NewHandler(service *Service, authMiddleware *middleware.AuthMiddleware) *Handler {
    return &Handler{
        service:        service,
        authMiddleware: authMiddleware,
    }
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
    router.HandleFunc("/search", h.authMiddleware.AuthHandler(h.handleSearch)).Methods("GET")
    router.HandleFunc("/search/workspaces", h.authMiddleware.AuthHandler(h.handleWorkspaceSearch)).Methods("GET")
    router.HandleFunc("/search/containers", h.authMiddleware.AuthHandler(h.handleContainerSearch)).Methods("GET")
    router.HandleFunc("/search/items", h.authMiddleware.AuthHandler(h.handleItemSearch)).Methods("GET")
    router.HandleFunc("/search/tags", h.authMiddleware.AuthHandler(h.handleTagSearch)).Methods("GET")
    router.HandleFunc("/search/containers/qr/{code}", h.authMiddleware.AuthHandler(h.handleContainerQRSearch)).Methods("GET")
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        writeError(w, http.StatusBadRequest, "search query is required")
        return
    }

    results, err := h.service.Search(query, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, results)
}

func (h *Handler) handleWorkspaceSearch(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        writeError(w, http.StatusBadRequest, "search query is required")
        return
    }

    results, err := h.service.SearchWorkspaces(query, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, results)
}

func (h *Handler) handleContainerSearch(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        writeError(w, http.StatusBadRequest, "search query is required")
        return
    }

    results, err := h.service.SearchContainers(query, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, results)
}

func (h *Handler) handleItemSearch(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        writeError(w, http.StatusBadRequest, "search query is required")
        return
    }

    results, err := h.service.SearchItems(query, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, results)
}

func (h *Handler) handleTagSearch(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        writeError(w, http.StatusBadRequest, "search query is required")
        return
    }

    results, err := h.service.SearchTags(query, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, results)
}

func (h *Handler) handleContainerQRSearch(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    qrCode := mux.Vars(r)["code"]
    if qrCode == "" {
        writeError(w, http.StatusBadRequest, "QR code is required")
        return
    }

    container, err := h.service.FindContainerByQR(qrCode, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, container)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}