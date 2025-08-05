package workspace

import (
	"encoding/json"
	"fmt"
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
    router.HandleFunc("/storage/workspaces", h.authMiddleware.ProfileAuthHandler(h.handleGetWorkspaces)).Methods("GET")
    router.HandleFunc("/storage/workspaces", h.authMiddleware.ProfileAuthHandler(h.handleCreateWorkspace)).Methods("POST")
    router.HandleFunc("/storage/workspaces/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetWorkspaceByID)).Methods("GET")
    router.HandleFunc("/storage/workspaces/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateWorkspace)).Methods("PUT")
    router.HandleFunc("/storage/workspaces/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteWorkspace)).Methods("DELETE")

    router.HandleFunc("/storage/workspaces/{id}/restore", h.authMiddleware.ProfileAuthHandler(h.handleRestoreWorkspace)).Methods("PUT")
}

func (h *Handler) handleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    workspaces, err := h.service.GetWorkspacesByFamilyID(profileCtx.FamilyID, profileCtx.ProfileID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, workspaces)
}

func (h *Handler) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    var req CreateWorkspaceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    workspace, err := h.service.CreateWorkspace(profileCtx.FamilyID, profileCtx.ProfileID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, workspace)
}

func (h *Handler) handleGetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    workspaceID, err := getWorkspaceIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    workspace, err := h.service.GetWorkspaceByID(workspaceID, profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    workspaceID, err := getWorkspaceIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req UpdateWorkspaceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    workspace, err := h.service.UpdateWorkspace(workspaceID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    workspaceID, err := getWorkspaceIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.DeleteWorkspace(workspaceID, profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]string{"deleted": string(workspaceID)})
}

func (h *Handler) handleRestoreWorkspace(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    workspaceID, err := getWorkspaceIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.RestoreWorkspace(workspaceID, profileCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]string{"restored": string(workspaceID)})
}

func getWorkspaceIDFromRequest(r *http.Request) (models.WorkspaceID, error) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    
    workspaceID := models.WorkspaceID(idStr)
    if !workspaceID.IsValid() {
        return "", fmt.Errorf("invalid workspace ID format")
    }
    
    return workspaceID, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}