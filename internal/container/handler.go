package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	router.HandleFunc("/containers", h.handleGetContainers).Methods("GET")
	router.HandleFunc("/containers", h.handleCreateContainer).Methods("POST")

	router.HandleFunc("/containers/{id}", h.handleGetContainerByID).Methods("GET")
	router.HandleFunc("/containers/{id}", h.handleDeleteContainer).Methods("DELETE")
	router.HandleFunc("/containers/{id}", h.handleUpdateContainer).Methods("PUT")

	router.HandleFunc("/containers/qr/{qrcode}", h.handleGetContainerByQR).Methods("GET")

}

func (h *Handler) handleGetContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.service.GetAllContainers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, containers)
}

func (h *Handler) handleGetContainerByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	container, err := h.service.GetContainerByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleGetContainerByQR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	qrCode := vars["qrcode"]

	if qrCode == "" {
		writeError(w, http.StatusBadRequest, "QR code parameter is required")
		return
	}

	qrCode = strings.TrimSpace(qrCode)
	if !strings.HasPrefix(qrCode, "STQRAGE-CONTAINER-") {
		writeError(w, http.StatusBadRequest, "invalid QR code format")
		return
	}

	container, err := h.service.GetContainerByQR(qrCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	var req CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	container, err := h.service.CreateContainer(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, container)
}

func (h *Handler) handleUpdateContainer(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	container, err := h.service.UpdateContainer(id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, container)
}

func (h *Handler) handleDeleteContainer(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.DeleteContainer(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func getIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return 0, fmt.Errorf("invalid id provided")
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
