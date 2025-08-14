package calendar

import (
	"encoding/json"
	"fmt"
	"net/http"
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
        writeError(w, http.StatusBadRequest, "invalid startTime format - expected RFC3339")
        return
    }

    endTime, err := time.Parse(time.RFC3339, query.Get("endTime"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid endTime format - expected RFC3339")
        return
    }
    
    if endTime.Before(startTime) {
        writeError(w, http.StatusBadRequest, "endTime must be after startTime")
        return
    }

    params := GetEventsParams{
        StartTime: startTime,
        EndTime:   endTime,
    }

    if assigneeIDsStr := query.Get("assigneeIds"); assigneeIDsStr != "" {
        var assigneeIDs []models.ProfileID
        for _, idStr := range strings.Split(assigneeIDsStr, ",") {
            idStr = strings.TrimSpace(idStr)
            profileID := models.ProfileID(idStr)
            if !profileID.IsValid() {
                writeError(w, http.StatusBadRequest, "invalid assigneeIds format")
                return
            }
            assigneeIDs = append(assigneeIDs, profileID)
        }
        params.AssigneeIDs = assigneeIDs
    }

    if sourceModulesStr := query.Get("sourceModules"); sourceModulesStr != "" {
        params.SourceModules = strings.Split(sourceModulesStr, ",")
    }

    if sourceIDStr := query.Get("sourceId"); sourceIDStr != "" {
        params.SourceID = &sourceIDStr
    }

    events, err := h.service.GetByDateRange(profileCtx.FamilyID, params, profileCtx)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    response := NormalizeEventsResponse(events)
    writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)

    var req CreateEventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    if err := h.validateCreateEventRequest(&req); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    event, err := h.service.Create(profileCtx, &req)  
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

    event, err := h.service.GetByID(eventID, profileCtx)
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

    if err := h.validateUpdateEventRequest(&req); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    req.UpdatedBy = profileCtx.ProfileID

    event, err := h.service.Update(eventID, &req, profileCtx)
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

    if err := h.validateUpdateInstanceRequest(&req); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    req.EventID = eventID
    req.UpdatedBy = profileCtx.ProfileID

    event, err := h.service.UpdateRecurringInstance(&req, profileCtx)
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

    if req.Date.IsZero() {
        writeError(w, http.StatusBadRequest, "date is required")
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

    writeJSON(w, http.StatusOK, map[string]string{"deleted": string(eventID)})
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

    writeJSON(w, http.StatusOK, map[string]string{"restored": string(eventID)})
}

func (h *Handler) validateCreateEventRequest(req *CreateEventRequest) error {
    if req.Title == "" {
        return fmt.Errorf("title is required")
    }
    
    if !req.AllDay && (req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime)) {
        return fmt.Errorf("endTime must be after startTime")
    }
    
    return nil
}

func (h *Handler) validateUpdateEventRequest(req *UpdateEventRequest) error {
    if req.Title == "" {
        return fmt.Errorf("title is required")
    }
    
    if !req.AllDay && (req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime)) {
        return fmt.Errorf("endTime must be after startTime")
    }
    
    return nil
}

func (h *Handler) validateUpdateInstanceRequest(req *UpdateRecurringInstanceRequest) error {
    if req.Title == nil && req.Description == nil && req.Location == nil && 
       req.StartTime == nil && req.EndTime == nil && req.AllDay == nil && req.AssigneeID == nil {
        return fmt.Errorf("at least one field must be updated")
    }

    if req.StartTime != nil && req.EndTime != nil && req.AllDay != nil && !*req.AllDay {
        if req.EndTime.Before(*req.StartTime) || req.EndTime.Equal(*req.StartTime) {
            return fmt.Errorf("endTime must be after startTime")
        }
    }
    
    return nil
}

func getIDFromRequest(r *http.Request) (models.EventID, error) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    
    eventID := models.EventID(idStr)
    if !eventID.IsValid() {
        return "", fmt.Errorf("invalid event ID format")
    }
    
    return eventID, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}