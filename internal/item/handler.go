package item

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chrisabs/storage/internal/middleware"
	"github.com/chrisabs/storage/internal/models"
	"github.com/gorilla/mux"
)

type ContainerService interface {
	GetContainerByID(id int) (*models.Container, error)
}

type Handler struct {
	service          *Service
	containerService ContainerService
	authMiddleware   *middleware.AuthMiddleware
}

func NewHandler(service *Service, containerService ContainerService, authMiddleware *middleware.AuthMiddleware) *Handler {
	return &Handler{
		service:          service,
		containerService: containerService,
		authMiddleware:   authMiddleware,
	}
}
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/items", h.authMiddleware.AuthHandler(h.handleGetItems)).Methods("GET")
	router.HandleFunc("/items", h.authMiddleware.AuthHandler(h.handleCreateItem)).Methods("POST")

	router.HandleFunc("/items/{id}", h.authMiddleware.AuthHandler(h.handleGetItem)).Methods("GET")
	router.HandleFunc("/items/{id}", h.authMiddleware.AuthHandler(h.handleUpdateItem)).Methods("PUT")
	router.HandleFunc("/items/{id}", h.authMiddleware.AuthHandler(h.handleDeleteItem)).Methods("DELETE")
}

func (h *Handler) handleGetItems(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	items, err := h.service.GetItemsByUserID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ContainerID != nil {
		container, err := h.containerService.GetContainerByID(*req.ContainerID)
		if err != nil {
			writeError(w, http.StatusNotFound, "container not found")
			return
		}
		if container.UserID != userID {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	item, err := h.service.CreateItem(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) handleGetItem(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	itemID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.service.GetItemByID(itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if item.ContainerID != nil {
		container, err := h.containerService.GetContainerByID(*item.ContainerID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if container.UserID != userID {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.Header.Get("UserId"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }

    itemID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req UpdateItemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if req.ContainerID != nil {
        container, err := h.containerService.GetContainerByID(*req.ContainerID)
        if err != nil {
            writeError(w, http.StatusNotFound, "container not found")
            return
        }
        if container.UserID != userID {
            writeError(w, http.StatusForbidden, "access denied")
            return
        }
    }

    item, err := h.service.UpdateItem(itemID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, item)
}

func (h *Handler) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("UserId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	itemID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.service.GetItemByID(itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if item.ContainerID != nil {
		container, err := h.containerService.GetContainerByID(*item.ContainerID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if container.UserID != userID {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	if err := h.service.DeleteItem(itemID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"deleted": itemID})
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
