package tag

import (
	"encoding/json"
	"net/http"
	"strconv"

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
    router.HandleFunc("/tags", h.authMiddleware.AuthHandler(h.handleGetTags)).Methods("GET")
    router.HandleFunc("/tags", h.authMiddleware.AuthHandler(h.handleCreateTag)).Methods("POST")
    router.HandleFunc("/tags/{id}", h.authMiddleware.AuthHandler(h.handleGetTag)).Methods("GET")
    router.HandleFunc("/tags/{id}", h.authMiddleware.AuthHandler(h.handleUpdateTag)).Methods("PUT")
    router.HandleFunc("/tags/{id}", h.authMiddleware.AuthHandler(h.handleDeleteTag)).Methods("DELETE")
    router.HandleFunc("/tags/assign", h.authMiddleware.AuthHandler(h.handleAssignTags)).Methods("POST")
}

func (h *Handler) handleGetTags(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    tags, err := h.service.GetAllTags(*userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, tags)
}

func (h *Handler) handleCreateTag(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    var req CreateTagRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    tag, err := h.service.CreateTag(*userCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, tag)
}

func (h *Handler) handleGetTag(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    id, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    tag, err := h.service.GetTagByID(id, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, tag)
}

func (h *Handler) handleUpdateTag(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    id, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req UpdateTagRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    tag, err := h.service.UpdateTag(id, *userCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, tag)
}

func (h *Handler) handleAssignTags(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    var req AssignTagsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := h.service.AssignTagsToItems(*userCtx.FamilyID, req.TagIDs, req.ItemIDs); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{"message": "tags assigned successfully"})
}

func (h *Handler) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    id, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.DeleteTag(id, *userCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"message": "tag deleted successfully"})
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