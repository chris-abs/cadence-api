package workspace

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
    router.HandleFunc("/workspaces", h.authMiddleware.AuthHandler(h.handleGetWorkspaces)).Methods("GET")
    router.HandleFunc("/workspaces", h.authMiddleware.AuthHandler(h.handleCreateWorkspace)).Methods("POST")
    router.HandleFunc("/workspaces/{id}", h.authMiddleware.AuthHandler(h.handleGetWorkspaceByID)).Methods("GET")
    router.HandleFunc("/workspaces/{id}", h.authMiddleware.AuthHandler(h.handleUpdateWorkspace)).Methods("PUT")
    router.HandleFunc("/workspaces/{id}", h.authMiddleware.AuthHandler(h.handleDeleteWorkspace)).Methods("DELETE")

    router.HandleFunc("/workspaces/{id}/restore", h.authMiddleware.AuthHandler(h.handleRestoreWorkspace)).Methods("PUT")
}

func (h *Handler) handleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)
    
    workspaces, err := h.service.GetWorkspacesByFamilyID(*userCtx.FamilyID, userCtx.profileId)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, workspaces)
}

func (h *Handler) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)

    var req CreateWorkspaceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    workspace, err := h.service.CreateWorkspace(userCtx, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, workspace)
}

func (h *Handler) handleGetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)
    
    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    workspace, err := h.service.GetWorkspaceByID(workspaceID, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)
    
    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req UpdateWorkspaceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    workspace, err := h.service.UpdateWorkspace(workspaceID, *userCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)
    
    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.DeleteWorkspace(workspaceID, *userCtx.FamilyID, userCtx.profileId); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]int{"deleted": workspaceID})
}

func (h *Handler) handleRestoreWorkspace(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.ProfileContext)
    
    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.RestoreWorkspace(workspaceID, *userCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]int{"restored": workspaceID})
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