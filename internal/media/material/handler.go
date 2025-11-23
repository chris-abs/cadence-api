package material

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrisabs/cadence/internal/cloud"
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
	router.HandleFunc("/media/material/summary", h.authMiddleware.ProfileAuthHandler(h.handleGetStatusSummary)).Methods("GET")
	router.HandleFunc("/media/material/enums", h.authMiddleware.ProfileAuthHandler(h.handleGetEnums)).Methods("GET")
	
	router.HandleFunc("/media/material", h.authMiddleware.ProfileAuthHandler(h.handleGetMedia)).Methods("GET")
	router.HandleFunc("/media/material", h.authMiddleware.ProfileAuthHandler(h.handleCreateMedia)).Methods("POST")
	router.HandleFunc("/media/material/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetMediaByID)).Methods("GET")
	router.HandleFunc("/media/material/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMedia)).Methods("PUT")
	router.HandleFunc("/media/material/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteMaterial)).Methods("DELETE")
	router.HandleFunc("/media/material/{id}/status", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMaterialStatus)).Methods("PATCH")
	router.HandleFunc("/media/material/{id}/review", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMaterialReview)).Methods("PATCH")
	router.HandleFunc("/media/material/{id}/poster", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMaterialPoster)).Methods("PATCH")
	router.HandleFunc("/media/material/{id}/poster", h.authMiddleware.ProfileAuthHandler(h.handleDeleteMaterialPoster)).Methods("DELETE")
}

func (h *Handler) handleGetMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	req := &MaterialSearchRequest{
		Query:     strings.TrimSpace(r.URL.Query().Get("query")),
		Type:      entities.MaterialType(r.URL.Query().Get("type")),
		Runtime:   entities.Runtime(r.URL.Query().Get("runtime")),
		WatchWith: entities.WatchWith(r.URL.Query().Get("watchWith")),
		Priority:  entities.Priority(r.URL.Query().Get("priority")),
		SortBy:    r.URL.Query().Get("sortBy"),
		Status:    entities.Status(r.URL.Query().Get("status")),
	}

	if profileIDStr := r.URL.Query().Get("profileId"); profileIDStr != "" {
		profileID := models.ProfileID(profileIDStr)
		if !profileID.IsValid() {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid profileId format: '%s' - must be a valid UUID", profileIDStr))
			return
		}
		req.ProfileID = &profileID
	}

	if sourceIDStr := r.URL.Query().Get("sourceId"); sourceIDStr != "" {
		sourceID := models.SourceID(sourceIDStr)
		// TODO: should we add validation for sourceID? would be beneficial when we allow users to create their own sources.
		req.SourceID = sourceID
	}

	if classificationIDStr := r.URL.Query().Get("classificationId"); classificationIDStr != "" {
		classificationID := models.ClassificationID(classificationIDStr)
		if !classificationID.IsValid() {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid classificationId format: '%s' - must be a valid UUID", classificationIDStr))
			return
		}
		req.ClassificationID = &classificationID
	}

	if includeClassifiedStr := r.URL.Query().Get("includeClassified"); includeClassifiedStr != "" {
		req.IncludeClassified = includeClassifiedStr == "true"
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

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		materialDataStr := r.FormValue("materialData")
		if materialDataStr == "" {
			writeError(w, http.StatusBadRequest, "missing materialData field")
			return
		}

		var req CreateMaterialRequest
		if err := json.Unmarshal([]byte(materialDataStr), &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid material data format")
			return
		}

		if file, header, err := r.FormFile("poster"); err == nil {
			defer file.Close()

			material, err := h.service.CreateMaterial(profileCtx.ProfileID, profileCtx.FamilyID, &req)
			if err != nil {
				if strings.Contains(err.Error(), "validation failed") {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}

			s3Handler, err := cloud.NewS3Handler()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to initialize storage")
				return
			}

			posterURL, err := s3Handler.UploadProfileMediaFile(header, profileCtx.FamilyID, profileCtx.ProfileID, "posters")
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to upload poster")
				return
			}

			updateReq := &UpdateMaterialRequest{
				Name:             material.Name,
				Type:             material.Type,
				Runtime:          material.Runtime,
				PosterURL:        posterURL,
				SourceIDs:        material.SourceIDs,
				ClassificationID: material.ClassificationID,
				WatchWith:        material.WatchWith,
				Status:           material.Status,
				Priority:         material.Priority,
				Notes:            material.Notes,
			}

			updatedMaterial, err := h.service.UpdateMaterial(material.ID, profileCtx.FamilyID, profileCtx.ProfileID, updateReq)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update material poster")
				return
			}

			writeJSON(w, http.StatusCreated, updatedMaterial)
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
		return
	}

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

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		materialDataStr := r.FormValue("materialData")
		if materialDataStr == "" {
			writeError(w, http.StatusBadRequest, "missing materialData field")
			return
		}

		var req UpdateMaterialRequest
		if err := json.Unmarshal([]byte(materialDataStr), &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid material data format")
			return
		}

		// Handle poster upload if present
		if file, header, err := r.FormFile("poster"); err == nil {
			defer file.Close()

			// Upload poster to S3
			s3Handler, err := cloud.NewS3Handler()
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to initialize storage")
				return
			}

			posterURL, err := s3Handler.UploadProfileMediaFile(header, profileCtx.FamilyID, profileCtx.ProfileID, "posters")
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to upload poster")
				return
			}

			// Update the request with the new poster URL
			req.PosterURL = posterURL
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
		return
	}

	// Handle JSON-only request (existing behavior)
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

func (h *Handler) handleUpdateMaterialReview(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateMaterialReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ReviewScore < 0.0 || req.ReviewScore > 10.0 {
		writeError(w, http.StatusBadRequest, "review score must be between 0.0 and 10.0")
		return
	}

	if err := h.service.UpdateMaterialReview(materialID, profileCtx.FamilyID, profileCtx.ProfileID, req.ReviewScore); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "review updated"})
}

func (h *Handler) handleUpdateMaterialPoster(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		writeError(w, http.StatusBadRequest, "poster update requires multipart form data")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("poster")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing poster file")
		return
	}
	defer file.Close()

	existingMaterial, err := h.service.GetMaterialByID(materialID, profileCtx.FamilyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if existingMaterial.ProfileID != profileCtx.ProfileID {
		writeError(w, http.StatusForbidden, "not authorized to update this material")
		return
	}

	s3Handler, err := cloud.NewS3Handler()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize storage")
		return
	}

	posterURL, err := s3Handler.UploadProfileMediaFile(header, profileCtx.FamilyID, profileCtx.ProfileID, "posters")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload poster")
		return
	}

	updateReq := &UpdateMaterialRequest{
		Name:             existingMaterial.Name,
		Type:             existingMaterial.Type,
		Runtime:          existingMaterial.Runtime,
		PosterURL:        posterURL,
		SourceIDs:        existingMaterial.SourceIDs,
		ClassificationID: existingMaterial.ClassificationID,
		WatchWith:        existingMaterial.WatchWith,
		Status:           existingMaterial.Status,
		Priority:         existingMaterial.Priority,
		Notes:            existingMaterial.Notes,
	}

	updatedMaterial, err := h.service.UpdateMaterial(materialID, profileCtx.FamilyID, profileCtx.ProfileID, updateReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update material poster")
		return
	}

	writeJSON(w, http.StatusOK, updatedMaterial)
}

func (h *Handler) handleDeleteMaterialPoster(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	materialID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existingMaterial, err := h.service.GetMaterialByID(materialID, profileCtx.FamilyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if existingMaterial.ProfileID != profileCtx.ProfileID {
		writeError(w, http.StatusForbidden, "not authorized to delete this material's poster")
		return
	}

	if existingMaterial.PosterURL != "" {
		s3Handler, err := cloud.NewS3Handler()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to initialize storage")
			return
		}

		if err := s3Handler.DeleteFileByURL(existingMaterial.PosterURL); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete poster")
			return
		}
	}

	updateReq := &UpdateMaterialRequest{
		Name:             existingMaterial.Name,
		Type:             existingMaterial.Type,
		Runtime:          existingMaterial.Runtime,
		PosterURL:        "", 
		SourceIDs:        existingMaterial.SourceIDs,
		ClassificationID: existingMaterial.ClassificationID,
		WatchWith:        existingMaterial.WatchWith,
		Status:           existingMaterial.Status,
		Priority:         existingMaterial.Priority,
		Notes:            existingMaterial.Notes,
	}

	updatedMaterial, err := h.service.UpdateMaterial(materialID, profileCtx.FamilyID, profileCtx.ProfileID, updateReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete material poster")
		return
	}

	writeJSON(w, http.StatusOK, updatedMaterial)
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
		return "", fmt.Errorf("invalid material ID format: '%s' - must be a valid UUID", idStr)
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