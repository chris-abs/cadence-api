package family

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
	router.HandleFunc("/family/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/family/login", h.handleLogin).Methods("POST")

	router.HandleFunc("/family", h.authMiddleware.FamilyAuthHandler(h.handleGetFamily)).Methods("GET")
	router.HandleFunc("/family", h.authMiddleware.FamilyAuthHandler(h.handleUpdateFamily)).Methods("PUT")
	
	router.HandleFunc("/family/modules", h.authMiddleware.ProfileAuthHandler(h.handleGetModules)).Methods("GET")
	router.HandleFunc("/family/modules/{moduleId}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateModule)).Methods("PUT")
	router.HandleFunc("/family/delete", h.authMiddleware.ProfileAuthHandler(h.handleDeleteFamily)).Methods("DELETE")
	router.HandleFunc("/family/restore", h.authMiddleware.ProfileAuthHandler(h.handleRestoreFamily)).Methods("PUT")
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Register(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Login(&req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleGetFamily(w http.ResponseWriter, r *http.Request) {
    familyCtx := r.Context().Value("family").(*models.FamilyContext)
    
    family, err := h.service.GetFamilyByID(familyCtx.FamilyID)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, family)
}

func (h *Handler) handleUpdateFamily(w http.ResponseWriter, r *http.Request) {
	familyCtx := r.Context().Value("family").(*models.FamilyContext)
	
	var req UpdateFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	family, err := h.service.UpdateFamily(familyCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, family)
}

func (h *Handler) handleGetModules(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	settings, err := h.service.GetFamilySettings(profileCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, settings.Modules)
}

func (h *Handler) handleUpdateModule(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	if !profileCtx.IsOwner && profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can update modules")
		return
	}
	
	vars := mux.Vars(r)
	moduleID := models.ModuleID(vars["moduleId"])
	
	var req UpdateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.ModuleID = moduleID
	
	if err := h.service.UpdateModule(profileCtx.FamilyID, &req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "module updated successfully"})
}

func (h *Handler) handleDeleteFamily(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	if !profileCtx.IsOwner {
		writeError(w, http.StatusForbidden, "only the family owner can delete the family")
		return
	}
	
	if err := h.service.DeleteFamily(profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "family deleted successfully"})
}

func (h *Handler) handleRestoreFamily(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	if !profileCtx.IsOwner {
		writeError(w, http.StatusForbidden, "only the family owner can restore the family")
		return
	}
	
	if err := h.service.RestoreFamily(profileCtx.FamilyID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "family restored successfully"})
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