package recent

import (
	"encoding/json"
	"net/http"

	"github.com/chrisabs/cadence/internal/middleware"
	"github.com/chrisabs/cadence/internal/models"
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
    router.HandleFunc("/recent", h.authMiddleware.AuthHandler(h.handleGetRecent)).Methods("GET")
}

func (h *Handler) handleGetRecent(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)
    
    if userCtx.FamilyID == nil {
        writeError(w, http.StatusBadRequest, "family ID is required")
        return
    }

    response, err := h.service.GetRecentEntities(*userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}