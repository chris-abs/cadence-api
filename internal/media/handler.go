package media

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/chrisabs/cadence/internal/media/entities"
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
	router.HandleFunc("/media/sources", h.authMiddleware.ProfileAuthHandler(h.handleGetSources)).Methods("GET")
	
	router.HandleFunc("/media", h.authMiddleware.ProfileAuthHandler(h.handleGetMedia)).Methods("GET")
	router.HandleFunc("/media", h.authMiddleware.ProfileAuthHandler(h.handleCreateMedia)).Methods("POST")
	router.HandleFunc("/media/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetMediaByID)).Methods("GET")
	router.HandleFunc("/media/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMedia)).Methods("PUT")
	router.HandleFunc("/media/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteMedia)).Methods("DELETE")
	router.HandleFunc("/media/{id}/status", h.authMiddleware.ProfileAuthHandler(h.handleUpdateMediaStatus)).Methods("PATCH")
}

func (h *Handler) handleGetMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	req := &MediaSearchRequest{
		Query:     strings.TrimSpace(r.URL.Query().Get("query")),
		Type:      entities.MediaType(r.URL.Query().Get("type")),
		Genre:     r.URL.Query().Get("genre"),
		WatchWith: entities.WatchWith(r.URL.Query().Get("watchWith")),
		Status:    entities.Status(r.URL.Query().Get("status")),
		Priority:  entities.Priority(r.URL.Query().Get("priority")),
	}

	if profileIDStr := r.URL.Query().Get("profileId"); profileIDStr != "" {
		profileID, err := strconv.Atoi(profileIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid profileId")
			return
		}
		req.ProfileID = &profileID
	}

	if sourceIDStr := r.URL.Query().Get("sourceId"); sourceIDStr != "" {
		sourceID, err := strconv.Atoi(sourceIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid sourceId")
			return
		}
		req.SourceID = sourceID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		req.Offset = offset
	}

	response, err := h.service.SearchMedia(profileCtx.FamilyID, profileCtx.ProfileID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCreateMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	var req CreateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	media, err := h.service.CreateMedia(profileCtx.ProfileID, profileCtx.FamilyID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, media)
}

func (h *Handler) handleGetMediaByID(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	mediaID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	media, err := h.service.GetMediaByID(mediaID, profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, media)
}

func (h *Handler) handleUpdateMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	mediaID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	media, err := h.service.UpdateMedia(mediaID, profileCtx.FamilyID, profileCtx.ProfileID, &req)
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

	writeJSON(w, http.StatusOK, media)
}

func (h *Handler) handleDeleteMedia(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	mediaID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.DeleteMedia(mediaID, profileCtx.FamilyID, profileCtx.ProfileID, profileCtx.ProfileID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"deleted": mediaID})
}

func (h *Handler) handleUpdateMediaStatus(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)

	mediaID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateMediaStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateMediaStatus(mediaID, profileCtx.FamilyID, profileCtx.ProfileID, req.Status); err != nil {
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
		requestedProfileID, err := strconv.Atoi(profileIDStr)
		if err != nil {
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

func (h *Handler) handleGetSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.service.GetAllSources()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string][]entities.Source{"sources": sources})
}

func getIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}