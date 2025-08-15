package sources

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/chrisabs/cadence/internal/media/sources/entities"
	"github.com/chrisabs/cadence/internal/media/sources/service"
	"github.com/chrisabs/cadence/internal/middleware"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/gorilla/mux"
)

type Handler struct {
	service        *service.Service
	authMiddleware *middleware.AuthMiddleware
}

func NewHandler(service *service.Service, authMiddleware *middleware.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/sources", h.authMiddleware.ProfileAuthHandler(h.handleGetSources)).Methods("GET")
	router.HandleFunc("/sources", h.authMiddleware.ProfileAuthHandler(h.handleCreateSource)).Methods("POST")
	router.HandleFunc("/sources/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetSource)).Methods("GET")
	router.HandleFunc("/sources/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateSource)).Methods("PUT")
	router.HandleFunc("/sources/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteSource)).Methods("DELETE")
}

func (h *Handler) handleGetSources(w http.ResponseWriter, r *http.Request) {
	var params entities.SourceSearchParams
	
	if category := r.URL.Query().Get("category"); category != "" {
		params.Category = &category
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = &limit
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = &offset
		}
	}
	
	sources, err := h.service.GetAllSources(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, sources)
}

func (h *Handler) handleGetSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := models.SourceID(vars["id"])
	
	if !sourceID.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid source ID")
		return
	}
	
	source, err := h.service.GetSourceByID(sourceID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, source)
}

func (h *Handler) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var req entities.CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	source, err := h.service.CreateSource(&req)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusCreated, source)
}

func (h *Handler) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := models.SourceID(vars["id"])
	
	if !sourceID.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid source ID")
		return
	}
	
	var req entities.UpdateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	source, err := h.service.UpdateSource(sourceID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "must be") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, source)
}

func (h *Handler) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	vars := mux.Vars(r)
	sourceID := models.SourceID(vars["id"])
	
	if !sourceID.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid source ID")
		return
	}
	
	if err := h.service.DeleteSource(sourceID, profileCtx.ProfileID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "being used") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"deleted": string(sourceID)})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}