package item

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chrisabs/cadence/internal/cloud"
	"github.com/chrisabs/cadence/internal/middleware"
	"github.com/chrisabs/cadence/internal/models"
	"github.com/chrisabs/cadence/internal/storage/entities"

	"github.com/gorilla/mux"
)

type ContainerService interface {
    GetContainerByID(id int, familyID int) (*entities.Container, error)
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
    router.HandleFunc("/items", h.authMiddleware.ProfileAuthHandler(h.handleGetItems)).Methods("GET")
    router.HandleFunc("/items", h.authMiddleware.ProfileAuthHandler(h.handleCreateItem)).Methods("POST")

    router.HandleFunc("/items/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetItem)).Methods("GET")
    router.HandleFunc("/items/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateItem)).Methods("PUT")
    router.HandleFunc("/items/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteItem)).Methods("DELETE")

    router.HandleFunc("/items/{id}/restore", h.authMiddleware.ProfileAuthHandler(h.handleRestoreItem)).Methods("PUT")
}

func (h *Handler) handleGetItems(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    items, err := h.service.GetItemsByFamilyID(profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, items)
}

func (h *Handler) handleCreateItem(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    var req CreateItemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if req.ContainerID != nil {
        if _, err := h.containerService.GetContainerByID(*req.ContainerID, profileCtx.FamilyID); err != nil {
            writeError(w, http.StatusNotFound, "container not found")
            return
        }
    }

    item, err := h.service.CreateItem(profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) handleGetItem(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    itemID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    item, err := h.service.GetItemByID(itemID, profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, item)
}

func (h *Handler) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
 
    itemID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid item ID")
        return
    }
 
	if _, err := h.service.GetItemByID(itemID, profileCtx.FamilyID); err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    var req UpdateItemRequest
    contentType := r.Header.Get("Content-Type")
    if strings.Contains(contentType, "multipart/form-data") {
        if err := r.ParseMultipartForm(10 << 20); err != nil {
            writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse form: %v", err))
            return
        }

        itemDataStr := r.FormValue("itemData")
        if itemDataStr == "" {
            writeError(w, http.StatusBadRequest, "missing itemData")
            return
        }

        if err := json.NewDecoder(strings.NewReader(itemDataStr)).Decode(&req); err != nil {
            writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid item data: %v", err))
            return
        }

        if files := r.MultipartForm.File["images"]; len(files) > 0 {
            s3Handler, err := cloud.NewS3Handler()
            if err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
            }

            for _, fileHeader := range files {
                url, err := s3Handler.UploadFile(fileHeader, fmt.Sprintf("items/%d", itemID))
                if err != nil {
                    writeError(w, http.StatusInternalServerError, err.Error())
                    return
                }

                if err := h.service.AddItemImage(itemID, profileCtx.FamilyID, url); err != nil {
                    writeError(w, http.StatusInternalServerError, err.Error())
                    return
                }
            }
        }
    } else {
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            writeError(w, http.StatusBadRequest, "invalid request body")
            return
        }
    }

    if len(req.ImagesToDelete) > 0 {
        for _, url := range req.ImagesToDelete {
            if err := h.service.DeleteItemImage(itemID, profileCtx.FamilyID, url); err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
            }
        }
    }

    updatedItem, err := h.service.UpdateItem(itemID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
 
    writeJSON(w, http.StatusOK, updatedItem)
}

func (h *Handler) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    itemID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.DeleteItem(itemID, profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]int{"deleted": itemID})
}

func (h *Handler) handleRestoreItem(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    itemID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.RestoreItem(itemID, profileCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, map[string]int{"restored": itemID})
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