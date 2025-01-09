package workspace

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chrisabs/storage/internal/middleware"
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
}

func (h *Handler) handleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.Header.Get("UserId"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }

    workspaces, err := h.service.GetWorkspacesByUserID(userID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, workspaces)
}

func (h *Handler) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.Header.Get("UserId"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }

    var req CreateWorkspaceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    workspace, err := h.service.CreateWorkspace(userID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, workspace)
}

func (h *Handler) handleGetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.Header.Get("UserId"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }

    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    workspace, err := h.service.GetWorkspaceByID(workspaceID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    if workspace.UserID != userID {
        writeError(w, http.StatusForbidden, "access denied")
        return
    }

    writeJSON(w, http.StatusOK, workspace)
}

func (h *Handler) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.Header.Get("UserId"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }

    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    workspace, err := h.service.GetWorkspaceByID(workspaceID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    if workspace.UserID != userID {
        writeError(w, http.StatusForbidden, "access denied")
        return
    }

    var req UpdateWorkspaceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    updatedWorkspace, err := h.service.UpdateWorkspace(workspaceID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, updatedWorkspace)
}

func (h *Handler) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.Header.Get("UserId"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }

    workspaceID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    workspace, err := h.service.GetWorkspaceByID(workspaceID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    if workspace.UserID != userID {
        writeError(w, http.StatusForbidden, "access denied")
        return
    }

    if err := h.service.DeleteWorkspace(workspaceID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]int{"deleted": workspaceID})
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