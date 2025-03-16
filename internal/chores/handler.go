package chores

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/chrisabs/cadence/internal/chores/entities"
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
	router.HandleFunc("/chores", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionRead)(h.handleGetChores)).Methods("GET")
	router.HandleFunc("/chores", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionWrite)(h.handleCreateChore)).Methods("POST")

	router.HandleFunc("/chores/{id}", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionRead)(h.handleGetChore)).Methods("GET")
	router.HandleFunc("/chores/{id}", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionWrite)(h.handleUpdateChore)).Methods("PUT")
	router.HandleFunc("/chores/{id}", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionWrite)(h.handleDeleteChore)).Methods("DELETE")

	router.HandleFunc("/chores/{id}/restore", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionWrite)(h.handleRestoreChore)).Methods("PUT")

	router.HandleFunc("/chores/instances", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionRead)(h.handleGetChoreInstances)).Methods("GET")

	router.HandleFunc("/chores/instances/{id}", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionRead)(h.handleGetChoreInstance)).Methods("GET")

	router.HandleFunc("/chores/instances/{id}/complete", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionWrite)(h.handleCompleteChoreInstance)).Methods("PUT")
	
	router.HandleFunc("/chores/verify-day", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionManage)(h.handleVerifyDay)).Methods("PUT")

	router.HandleFunc("/chores/instances/{id}/review", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionManage)(h.handleReviewChore)).Methods("PUT")
	
	router.HandleFunc("/chores/daily-verification", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionRead)(h.handleGetDailyVerification)).Methods("GET")

	router.HandleFunc("/chores/stats", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionRead)(h.handleGetChoreStats)).Methods("GET")

	router.HandleFunc("/chores/generate", h.authMiddleware.ModuleMiddleware(models.ModuleChores, models.PermissionManage)(h.handleGenerateChoreInstances)).Methods("POST")
}

func (h *Handler) handleGetChores(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	assigneeIDStr := r.URL.Query().Get("assigneeId")
	
	var chores []*entities.Chore
	var err error
	
	if assigneeIDStr != "" {
		assigneeID, err := strconv.Atoi(assigneeIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid assignee ID")
			return
		}
		
		chores, err = h.service.GetChoresByAssigneeID(assigneeID, profileCtx.FamilyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		chores, err = h.service.GetChoresByFamilyID(profileCtx.FamilyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, chores)
}

func (h *Handler) handleCreateChore(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	var req CreateChoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	chore, err := h.service.CreateChore(profileCtx.ProfileID, profileCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusCreated, chore)
}

func (h *Handler) handleGetChore(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	chore, err := h.service.GetChoreByID(id, profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, chore)
}

func (h *Handler) handleUpdateChore(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	var req UpdateChoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	chore, err := h.service.UpdateChore(id, profileCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, chore)
}

func (h *Handler) handleDeleteChore(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    id, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }
    
    if err := h.service.DeleteChore(id, profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]string{"message": "chore deleted successfully"})
}

func (h *Handler) handleRestoreChore(w http.ResponseWriter, r *http.Request) {
    profileCtx := r.Context().Value("profile").(*models.ProfileContext)
    
    id, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }
    
    if err := h.service.RestoreChore(id, profileCtx.FamilyID); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]string{"message": "chore restored successfully"})
}

func (h *Handler) handleGetChoreInstances(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	dateStr := r.URL.Query().Get("date")
	assigneeIDStr := r.URL.Query().Get("assigneeId")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	
	if dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format (use YYYY-MM-DD)")
			return
		}
		
		instances, err := h.service.GetInstancesByDueDate(date, profileCtx.FamilyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		writeJSON(w, http.StatusOK, instances)
		return
	}
	
	if assigneeIDStr != "" && startDateStr != "" && endDateStr != "" {
		assigneeID, err := strconv.Atoi(assigneeIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid assignee ID")
			return
		}
		
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start date format (use YYYY-MM-DD)")
			return
		}
		
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end date format (use YYYY-MM-DD)")
			return
		}
		
		instances, err := h.service.GetInstancesByAssignee(assigneeID, profileCtx.FamilyID, startDate, endDate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		writeJSON(w, http.StatusOK, instances)
		return
	}
	
	today := time.Now().UTC().Truncate(24 * time.Hour)
	instances, err := h.service.GetInstancesByDueDate(today, profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, instances)
}

func (h *Handler) handleGetChoreInstance(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	instance, err := h.service.GetInstanceByID(id, profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, instance)
}

func (h *Handler) handleCompleteChoreInstance(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	var req UpdateChoreInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	instance, err := h.service.CompleteChoreInstance(id, profileCtx.ProfileID, profileCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, instance)
}

func (h *Handler) handleVerifyDay(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	if profileCtx.Role == nil || *profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can verify chores")
		return
	}
	
	var req VerifyDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	if err := h.service.VerifyDay(profileCtx.ProfileID, profileCtx.FamilyID, &req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "day verified successfully"})
}

func (h *Handler) handleReviewChore(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	if profileCtx.Role == nil || *profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can review chores")
		return
	}
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	var req ReviewChoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	instance, err := h.service.ReviewChore(id, profileCtx.ProfileID, profileCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, instance)
}

func (h *Handler) handleGetDailyVerification(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	dateStr := r.URL.Query().Get("date")
	assigneeIDStr := r.URL.Query().Get("assigneeId")
	
	if dateStr == "" || assigneeIDStr == "" {
		writeError(w, http.StatusBadRequest, "date and assigneeId are required")
		return
	}
	
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format")
		return
	}
	
	assigneeID, err := strconv.Atoi(assigneeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid assigneeId")
		return
	}
	
	verification, err := h.service.GetDailyVerification(date, assigneeID, profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, verification)
}

func (h *Handler) handleGetChoreStats(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	profileIdStr := r.URL.Query().Get("profileId")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	
	if startDateStr == "" || endDateStr == "" {
		writeError(w, http.StatusBadRequest, "startDate and endDate are required")
		return
	}
	
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startDate format (use YYYY-MM-DD)")
		return
	}
	
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid endDate format (use YYYY-MM-DD)")
		return
	}
	
	var profileId int
	if profileIdStr != "" {
		profileId, err = strconv.Atoi(profileIdStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid profileId")
			return
		}
	} else {
		profileId = profileCtx.ProfileID
	}
	
	stats, err := h.service.GetChoreStats(profileId, profileCtx.FamilyID, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) handleGenerateChoreInstances(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	if profileCtx.Role == nil || *profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can generate chore instances")
		return
	}
	
	if err := h.service.GenerateDailyChoreInstances(profileCtx.FamilyID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "chore instances generated successfully"})
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