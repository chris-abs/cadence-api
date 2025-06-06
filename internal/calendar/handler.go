package calendar

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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
    router.HandleFunc("/calendar/events/{id}/modify-instance", h.authMiddleware.ProfileAuthHandler(h.handleUpdateInstance)).Methods("POST")
    router.HandleFunc("/calendar/events/{id}/cancel-instance", h.authMiddleware.ProfileAuthHandler(h.handleCancelInstance)).Methods("POST")
}

func (h *Handler) handleGetEvents(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    query := r.URL.Query()

    startTime, err := time.Parse(time.RFC3339, query.Get("startTime"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid startTime format")
        return
    }

    endTime, err := time.Parse(time.RFC3339, query.Get("endTime"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid endTime format")
        return
    }

    params := GetEventsParams{
        StartTime: startTime,
        EndTime:   endTime,
    }

    if assigneeIDsStr := query.Get("assigneeIds"); assigneeIDsStr != "" {
        assigneeIDs := []int{}
        for _, idStr := range strings.Split(assigneeIDsStr, ",") {
            id, err := strconv.Atoi(idStr)
            if err != nil {
                writeError(w, http.StatusBadRequest, "invalid assigneeIds format")
                return
            }
            assigneeIDs = append(assigneeIDs, id)
        }
        params.AssigneeIDs = assigneeIDs
    }

    if sourceModulesStr := query.Get("sourceModules"); sourceModulesStr != "" {
        params.SourceModules = strings.Split(sourceModulesStr, ",")
    }

    if sourceIDStr := query.Get("sourceId"); sourceIDStr != "" {
        sourceID, err := strconv.Atoi(sourceIDStr)
        if err != nil {
            writeError(w, http.StatusBadRequest, "invalid sourceId format")
            return
        }
        params.SourceID = &sourceID
    }

    events, err := h.service.GetByDateRange(profileCtx.FamilyID, params, profileCtx.ProfileID)
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

    event, err := h.service.GetByID(eventID, profileCtx.FamilyID, profileCtx.ProfileID)
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

    req.UpdatedBy = profileCtx.ProfileID

    event, err := h.service.Update(eventID, profileCtx.FamilyID, &req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, event)
}

func (h *Handler) handleUpdateInstance(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    eventID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid event ID")
        return
    }

    var req UpdateRecurringInstanceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    req.EventID = eventID
    req.UpdatedBy = profileCtx.ProfileID

    event, err := h.service.UpdateRecurringInstance(&req, profileCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, event)
}

func (h *Handler) handleCancelInstance(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    eventID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req struct {
        Date time.Time `json:"date"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := h.service.CancelRecurringInstance(eventID, profileCtx.FamilyID, req.Date, profileCtx.ProfileID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
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