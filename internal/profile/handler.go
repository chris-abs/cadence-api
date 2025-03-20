package profile

import (
	"encoding/json"
	"mime/multipart"
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
	router.HandleFunc("/profiles", h.authMiddleware.FamilyAuthHandler(h.handleGetProfiles)).Methods("GET")
	router.HandleFunc("/profiles", h.authMiddleware.FamilyAuthHandler(h.handleCreateProfile)).Methods("POST")
	router.HandleFunc("/profiles/select", h.authMiddleware.FamilyAuthHandler(h.handleSelectProfile)).Methods("POST")
	router.HandleFunc("/profiles/verify", h.authMiddleware.FamilyAuthHandler(h.handleVerifyPin)).Methods("POST")

	router.HandleFunc("/profiles/{id}", h.authMiddleware.ProfileAuthHandler(h.handleGetProfile)).Methods("GET")
	router.HandleFunc("/profiles/{id}", h.authMiddleware.ProfileAuthHandler(h.handleUpdateProfile)).Methods("PUT")
	router.HandleFunc("/profiles/{id}", h.authMiddleware.ProfileAuthHandler(h.handleDeleteProfile)).Methods("DELETE")
	router.HandleFunc("/profiles/{id}/restore", h.authMiddleware.ProfileAuthHandler(h.handleRestoreProfile)).Methods("PUT")
}

func (h *Handler) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	familyCtx := r.Context().Value("family").(*models.FamilyContext)
	
	profiles, err := h.service.GetProfilesByFamilyID(familyCtx.FamilyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, &ProfilesList{Profiles: profiles})
}

func (h *Handler) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	familyCtx := r.Context().Value("family").(*models.FamilyContext)
	
	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	profile, err := h.service.CreateProfile(familyCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusCreated, profile)
}

func (h *Handler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	profile, err := h.service.GetProfileByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	if profile.FamilyID != profileCtx.FamilyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	
	writeJSON(w, http.StatusOK, profile)
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !profileCtx.IsOwner && profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can update profiles")
		return
	}
	
	var imageFile *multipart.FileHeader
	if err := r.ParseMultipartForm(10 << 20); err == nil {
		if file, header, err := r.FormFile("image"); err == nil {
			defer file.Close()
			imageFile = header
		}
	}
	
	var req UpdateProfileRequest
	profileData := r.FormValue("profileData")
	if profileData != "" {
		if err := json.Unmarshal([]byte(profileData), &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid profile data")
			return
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}
	
	profile, err := h.service.UpdateProfile(id, profileCtx.FamilyID, &req, imageFile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, profile)
}

func (h *Handler) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	if !profileCtx.IsOwner && profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can delete profiles")
		return
	}
	
	if err := h.service.DeleteProfile(id, profileCtx.FamilyID, profileCtx.ProfileID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "profile deleted successfully"})
}

func (h *Handler) handleRestoreProfile(w http.ResponseWriter, r *http.Request) {
	profileCtx := r.Context().Value("profile").(*models.ProfileContext)
	
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	if !profileCtx.IsOwner && profileCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can restore profiles")
		return
	}
	
	if err := h.service.RestoreProfile(id, profileCtx.FamilyID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"message": "profile restored successfully"})
}

func (h *Handler) handleSelectProfile(w http.ResponseWriter, r *http.Request) {
	familyCtx := r.Context().Value("family").(*models.FamilyContext)
	
	var req SelectProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	profileResponse, err := h.service.SelectProfile(familyCtx.FamilyID, &req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, profileResponse)
}

func (h *Handler) handleVerifyPin(w http.ResponseWriter, r *http.Request) {
    familyCtx := r.Context().Value("family").(*models.FamilyContext)
    
    var req VerifyPinRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    profileResponse, err := h.service.VerifyPin(familyCtx.FamilyID, req.ProfileID, req.Pin)
    if err != nil {
        if strings.Contains(err.Error(), "invalid PIN") {
            writeError(w, http.StatusForbidden, "Invalid PIN")
            return
        }
        
        writeError(w, http.StatusUnauthorized, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, profileResponse)
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