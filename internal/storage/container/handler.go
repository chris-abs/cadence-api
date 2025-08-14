package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
    router.HandleFunc("/storage/containers", h.authMiddleware.ProfileAuthHandler(h.handleGetContainers)).Methods("GET")
    router.HandleFunc("/storage/containers", h.authMiddleware.ProfileAuthHandler(h.handleCreateContainer)).Methods("POST")

    router.HandleFunc("/storage/containers/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetContainerByID)).Methods("GET")
    router.HandleFunc("/storage/containers/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteContainer)).Methods("DELETE")
    router.HandleFunc("/storage/containers/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateContainer)).Methods("PUT")
    
    router.HandleFunc("/storage/containers/{id}/restore", h.authMiddleware.ProfileAuthHandler(h.handleRestoreContainer)).Methods("PUT")

    router.HandleFunc("/storage/containers/qr/{qrcode}", h.authMiddleware.ProfileAuthHandler(h.handleGetContainerByQR)).Methods("GET")
}

func (h *Handler) handleGetContainers(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    containers, err := h.service.GetContainersByFamilyID(profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, containers)
}

func (h *Handler) handleCreateContainer(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    var req CreateContainerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    container, err := h.service.CreateContainer(profileCtx.ProfileID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, container)
}

func (h *Handler) handleGetContainerByID(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    containerID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    container, err := h.service.GetContainerByID(containerID, profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleUpdateContainer(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    containerID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req UpdateContainerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    container, err := h.service.UpdateContainer(containerID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleGetContainerByQR(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    vars := mux.Vars(r)
    qrCode := strings.TrimSpace(vars["qrcode"])
    if qrCode == "" {
        writeError(w, http.StatusBadRequest, "QR code is required")
        return
    }

    container, err := h.service.GetContainerByQR(qrCode, profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleDeleteContainer(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    containerID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.DeleteContainer(containerID, profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"deleted": containerID.String()})
}

func (h *Handler) handleRestoreContainer(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    containerID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.RestoreContainer(containerID, profileCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"restored": containerID.String()})
}

func getIDFromRequest(r *http.Request) (models.ContainerID, error) {
    vars := mux.Vars(r)
    id := models.ContainerID(vars["id"])
    if !id.IsValid() {
        return "", fmt.Errorf("invalid container ID format")
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