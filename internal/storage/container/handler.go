package container

import (
	"encoding/json"
	"net/http"
	"strconv"
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
    router.HandleFunc("/containers", h.authMiddleware.AuthHandler(h.handleGetContainers)).Methods("GET")
    router.HandleFunc("/containers", h.authMiddleware.AuthHandler(h.handleCreateContainer)).Methods("POST")
    router.HandleFunc("/containers/{id}", h.authMiddleware.AuthHandler(h.handleGetContainerByID)).Methods("GET")
    router.HandleFunc("/containers/{id}", h.authMiddleware.AuthHandler(h.handleDeleteContainer)).Methods("DELETE")
    router.HandleFunc("/containers/{id}", h.authMiddleware.AuthHandler(h.handleUpdateContainer)).Methods("PUT")
    router.HandleFunc("/containers/qr/{qrcode}", h.authMiddleware.AuthHandler(h.handleGetContainerByQR)).Methods("GET")
}

func (h *Handler) handleGetContainers(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    containers, err := h.service.GetContainersByFamilyID(*userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, containers)
}

func (h *Handler) handleCreateContainer(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    var req CreateContainerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    container, err := h.service.CreateContainer(userCtx.UserID, *userCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, container)
}

func (h *Handler) handleGetContainerByID(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    containerID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    container, err := h.service.GetContainerByID(containerID, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleUpdateContainer(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

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

    container, err := h.service.UpdateContainer(containerID, *userCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleDeleteContainer(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    containerID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.DeleteContainer(containerID, *userCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]int{"deleted": containerID})
}

func (h *Handler) handleGetContainerByQR(w http.ResponseWriter, r *http.Request) {
    userCtx := r.Context().Value("user").(*models.UserContext)

    vars := mux.Vars(r)
    qrCode := strings.TrimSpace(vars["qrcode"])
    if qrCode == "" {
        writeError(w, http.StatusBadRequest, "QR code is required")
        return
    }

    container, err := h.service.GetContainerByQR(qrCode, *userCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, container)
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
