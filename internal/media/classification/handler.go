package classification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chrisabs/cadence/internal/cloud"
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
	router.HandleFunc("/classifications", h.authMiddleware.ProfileAuthHandler(h.handleCreateClassification)).Methods("POST")
	router.HandleFunc("/classifications", h.authMiddleware.ProfileAuthHandler(h.handleGetAllClassifications)).Methods("GET")
	router.HandleFunc("/classifications/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetClassification)).Methods("GET")
	router.HandleFunc("/classifications/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateClassification)).Methods("PUT")
	router.HandleFunc("/classifications/{id}/image", h.authMiddleware.ProfileAuthHandler(h.handleUploadClassificationImage)).Methods("POST")
	router.HandleFunc("/classifications/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteClassification)).Methods("DELETE")
}

func (h *Handler) handleCreateClassification(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	var req CreateClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	classification, err := h.service.CreateClassification(&req, profileCtx.FamilyID, profileCtx.ProfileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	writeJSON(w, http.StatusCreated, classification)
}

func (h *Handler) handleGetAllClassifications(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	var limit, offset *int
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = &l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = &o
		}
	}
	
	params := ClassificationSearchRequest{
		FamilyID: profileCtx.FamilyID,
		Limit:    limit,
		Offset:   offset,
	}
	
	classifications, err := h.service.GetAllClassifications(profileCtx.FamilyID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, classifications)
}

func (h *Handler) handleGetClassification(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	classification, err := h.service.GetClassificationByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	if classification.FamilyID != profileCtx.FamilyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	
	writeJSON(w, http.StatusOK, classification)
}

func (h *Handler) handleUpdateClassification(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	existingClassification, err := h.service.GetClassificationByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	if existingClassification.FamilyID != profileCtx.FamilyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	
	var req UpdateClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	classification, err := h.service.UpdateClassification(id, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, classification)
}

func (h *Handler) handleUploadClassificationImage(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}
	
	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()
	
	s3Handler, err := cloud.NewS3Handler()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize storage")
		return
	}
	
	imageURL, err := s3Handler.UploadProfileMediaFile(header, profileCtx.FamilyID, profileCtx.ProfileID, "classifications")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}
	
	classification, err := h.service.UpdateClassificationImage(id, imageURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update classification image")
		return
	}
	
	writeJSON(w, http.StatusOK, classification)
}

func (h *Handler) handleDeleteClassification(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	if err := h.service.DeleteClassification(id, profileCtx.ProfileID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"deleted": id.String()})
}

func getIDFromRequest(r *http.Request) (models.ClassificationID, error) {
	vars := mux.Vars(r)
	id := models.ClassificationID(vars["id"])
	if !id.IsValid() {
		return "", fmt.Errorf("invalid classification ID format")
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
