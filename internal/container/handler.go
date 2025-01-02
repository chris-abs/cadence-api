package container

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/chrisabs/storage/internal/middleware"
	"github.com/gorilla/mux"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/containers", middleware.AuthMiddleware(h.handleGetContainers)).Methods("GET")
	router.HandleFunc("/containers", middleware.AuthMiddleware(h.handleCreateContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}", middleware.AuthMiddleware(h.handleGetContainerByID)).Methods("GET")
	router.HandleFunc("/containers/{id}", middleware.AuthMiddleware(h.handleDeleteContainer)).Methods("DELETE")
	router.HandleFunc("/containers/{id}", middleware.AuthMiddleware(h.handleUpdateContainer)).Methods("PUT")
	router.HandleFunc("/containers/qr/{qrcode}", middleware.AuthMiddleware(h.handleGetContainerByQR)).Methods("GET")
}

func (h *Handler) handleGetContainers(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	containers, err := h.service.GetContainersByUserID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, containers)
}

func (h *Handler) handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	container, err := h.service.CreateContainer(userID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, container)
}

func (h *Handler) handleGetContainerByID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	containerID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	container, err := h.service.GetContainerByID(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if container.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleUpdateContainer(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	containerID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	container, err := h.service.GetContainerByID(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if container.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var req UpdateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedContainer, err := h.service.UpdateContainer(containerID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updatedContainer)
}

func (h *Handler) handleDeleteContainer(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	containerID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	container, err := h.service.GetContainerByID(containerID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if container.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	if err := h.service.DeleteContainer(containerID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"deleted": containerID})
}

func (h *Handler) handleGetContainerByQR(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	vars := mux.Vars(r)
	qrCode := strings.TrimSpace(vars["qrcode"])
	if qrCode == "" {
		writeError(w, http.StatusBadRequest, "QR code is required")
		return
	}

	container, err := h.service.GetContainerByQR(qrCode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if container.UserID != userID {
		writeError(w, http.StatusForbidden, "access denied")
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
