package calendar

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
    router.HandleFunc("/calendar/events", h.authMiddleware.ProfileAuthHandler(h.handleGetEvents)).Methods("GET")
    router.HandleFunc("/calendar/events", h.authMiddleware.ProfileAuthHandler(h.handleCreateEvent)).Methods("POST")
    
    router.HandleFunc("/calendar/events/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetEvent)).Methods("GET")
    router.HandleFunc("/calendar/events/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateEvent)).Methods("PUT")
    router.HandleFunc("/calendar/events/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteEvent)).Methods("DELETE")
    
    router.HandleFunc("/calendar/events/{id}/restore", h.authMiddleware.ProfileAuthHandler(h.handleRestoreEvent)).Methods("PUT")
}

func (h *Handler) handleGetEvents(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    var params GetEventsParams
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request parameters")
        return
    }

    events, err := h.service.GetByDateRange(profileCtx.FamilyID, params)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, events)
}

func (h *Handler) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    var req CreateEventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    event, err := h.service.Create(profileCtx.ProfileID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusCreated, event)
}

func (h *Handler) handleGetEvent(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    eventID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    event, err := h.service.GetByID(eventID, profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, event)
}

func (h *Handler) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    eventID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid event ID")
        return
    }

    var req UpdateEventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    event, err := h.service.Update(eventID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, event)
}

func (h *Handler) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    eventID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.Delete(eventID, profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, map[string]int{"deleted": eventID})
}

func (h *Handler) handleRestoreEvent(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    eventID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    if err := h.service.RestoreDeleted(eventID, profileCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, map[string]int{"restored": eventID})
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