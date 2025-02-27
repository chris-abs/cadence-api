package membership

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chrisabs/storage/internal/middleware"
	"github.com/chrisabs/storage/internal/models"
	"github.com/gorilla/mux"
)

type Handler struct {
	service        *Service
	authMiddleware *middleware.AuthMiddleware
}

type UpdateMembershipRequest struct {
	Role    models.UserRole `json:"role"`
	IsOwner bool            `json:"isOwner"`
}

func NewHandler(service *Service, authMiddleware *middleware.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/family-memberships", h.authMiddleware.AuthHandler(h.handleGetUserMemberships)).Methods("GET")
	router.HandleFunc("/family-memberships/active", h.authMiddleware.AuthHandler(h.handleGetActiveMembership)).Methods("GET")
	router.HandleFunc("/family-memberships/{id}", h.authMiddleware.AuthHandler(h.handleUpdateMembership)).Methods("PUT")
	router.HandleFunc("/family-memberships/{id}", h.authMiddleware.AuthHandler(h.handleDeleteMembership)).Methods("DELETE")
	
	router.HandleFunc("/families/{familyId}/memberships", h.authMiddleware.AuthHandler(h.handleGetFamilyMemberships)).Methods("GET")
}

func (h *Handler) handleGetUserMemberships(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value("user").(*models.UserContext)
	
	memberships, err := h.service.GetMembershipsByUserID(userCtx.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, memberships)
}

func (h *Handler) handleGetActiveMembership(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value("user").(*models.UserContext)
	
	membership, err := h.service.GetActiveMembershipForUser(userCtx.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, membership)
}

func (h *Handler) handleGetFamilyMemberships(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value("user").(*models.UserContext)
	
	// User must be a member of the family to view its memberships
	if userCtx.FamilyID == nil {
		writeError(w, http.StatusForbidden, "user is not part of a family")
		return
	}
	
	vars := mux.Vars(r)
	familyIDStr := vars["familyId"]
	familyID, err := strconv.Atoi(familyIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid family ID")
		return
	}
	
	// User can only view memberships for their own family
	if *userCtx.FamilyID != familyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	
	// Only allow parents to view all memberships
	if userCtx.Role == nil || *userCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can view family memberships")
		return
	}
	
	memberships, err := h.service.GetMembershipsByFamilyID(familyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, memberships)
}

func (h *Handler) handleUpdateMembership(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value("user").(*models.UserContext)
	
	// Check if user is a parent
	if userCtx.Role == nil || *userCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can update memberships")
		return
	}
	
	// Get membership ID from URL
	vars := mux.Vars(r)
	membershipIDStr := vars["id"]
	membershipID, err := strconv.Atoi(membershipIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid membership ID")
		return
	}
	
	// Get membership details to check if it's in the user's family
	membership, err := h.service.GetMembershipByID(membershipID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	// User can only update memberships in their own family
	if userCtx.FamilyID == nil || membership.FamilyID != *userCtx.FamilyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	
	// Parse request body
	var req UpdateMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	// Update membership
	updatedMembership, err := h.service.UpdateMembership(membershipID, req.Role, req.IsOwner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedMembership)
}

func (h *Handler) handleDeleteMembership(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value("user").(*models.UserContext)
	
	// Check if user is a parent
	if userCtx.Role == nil || *userCtx.Role != models.RoleParent {
		writeError(w, http.StatusForbidden, "only parents can delete memberships")
		return
	}
	
	// Get membership ID from URL
	vars := mux.Vars(r)
	membershipIDStr := vars["id"]
	membershipID, err := strconv.Atoi(membershipIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid membership ID")
		return
	}
	
	// Get membership details to check if it's in the user's family
	membership, err := h.service.GetMembershipByID(membershipID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	// User can only delete memberships in their own family
	if userCtx.FamilyID == nil || membership.FamilyID != *userCtx.FamilyID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}
	
	// Don't allow deleting the owner's membership
	if membership.IsOwner {
		writeError(w, http.StatusForbidden, "cannot delete the family owner's membership")
		return
	}
	
	// Delete membership
	if err := h.service.DeleteMembership(membershipID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "membership deleted successfully"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}