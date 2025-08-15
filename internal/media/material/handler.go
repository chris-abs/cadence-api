package material

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrisabs/cadence/internal/media/material/entities"
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
	router.HandleFunc("/media/summary", h.authMiddleware.ProfileAuthHandler(h.handleGetStatusSummary)).Methods("GET")
	router.HandleFunc("/media/enums", h.authMiddleware.ProfileAuthHandler(h.handleGetEnums)).Methods("GET")
	
	router.HandleFunc("/media", h.authMiddleware.ProfileAuthHandler(h.handleGetMedia)).Methods("GET")
	router.HandleFunc("/media", h.authMiddleware.ProfileAuthHandler(h.handleCreateMedia)).Methods("POST")
	router.HandleFunc("/media/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetMediaByID)).Methods("GET")
	router.HandleFunc("/media/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMedia)).Methods("PUT")
	router.HandleFunc("/media/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteMaterial)).Methods("DELETE")
	router.HandleFunc("/media/{id}/status", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMaterialStatus)).Methods("PATCH")
}

func (h *Handler) handleGetMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	req := &MaterialSearchRequest{
		Query:     strings.TrimSpace(r.URL.Query().Get("query")),
		Type:      entities.MaterialType(r.URL.Query().Get("type")),
		Genre:     r.URL.Query().Get("genre"),
		WatchWith: entities.WatchWith(r.URL.Query().Get("watchWith")),
		Status:    entities.Status(r.URL.Query().Get("status")),
		Priority:  entities.Priority(r.URL.Query().Get("priority")),
	}

	if profileIDStr := r.URL.Query().Get("profileId"); profileIDStr != "" {
		profileID := models.ProfileID(profileIDStr)
		if !profileID.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid profileId")
			return
		}
		req.ProfileID = &profileID
	}

	if sourceIDStr := r.URL.Query().Get("sourceId"); sourceIDStr != "" {
		sourceID := models.SourceID(sourceIDStr)
		// TODO: should we add validation for sourceID? would be beneficial when we allow users to create their own sources.
		req.SourceID = sourceID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var limit int
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit < 0 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		var offset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil || offset < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		req.Offset = offset
	}

	response, err := h.service.SearchMaterial(profileCtx.FamilyID, profileCtx.ProfileID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCreateMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	var req CreateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	material, err := h.service.CreateMaterial(profileCtx.ProfileID, profileCtx.FamilyID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, material)
}

func (h *Handler) handleGetMediaByID(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	material, err := h.service.GetMaterialByID(materialID, profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, material)
}

func (h *Handler) handleUpdateMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	material, err := h.service.UpdateMaterial(materialID, profileCtx.FamilyID, profileCtx.ProfileID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, material)
}

func (h *Handler) handleDeleteMaterial(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.DeleteMaterial(materialID, profileCtx.FamilyID, profileCtx.ProfileID, profileCtx.ProfileID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": string(materialID)})
}

func (h *Handler) handleUpdateMaterialStatus(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateMaterialStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateMaterialStatus(materialID, profileCtx.FamilyID, profileCtx.ProfileID, req.Status); err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) handleGetStatusSummary(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	profileID := profileCtx.ProfileID
	if profileIDStr := r.URL.Query().Get("profileId"); profileIDStr != "" {
		requestedProfileID := models.ProfileID(profileIDStr)
		if !requestedProfileID.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid profileId")
			return
		}
		profileID = requestedProfileID
	}

	summary, err := h.service.GetStatusSummary(profileCtx.FamilyID, profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) handleGetEnums(w http.ResponseWriter, r *http.Request) {
	enums, err := h.service.GetEnums()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, enums)
}


func getIDFromRequest(r *http.Request) (models.MaterialID, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	materialID := models.MaterialID(idStr)
	if !materialID.IsValid() {
		return "", fmt.Errorf("invalid material ID format")
	}
	
	return materialID, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}