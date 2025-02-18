package family

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

func NewHandler(service *Service, authMiddleware *middleware.AuthMiddleware) *Handler {
    return &Handler{
        service:        service,
        authMiddleware: authMiddleware,
    }
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
    router.HandleFunc("/families", h.authMiddleware.AuthHandler(h.handleCreateFamily)).Methods("POST")
    router.HandleFunc("/families/{id}", h.authMiddleware.AuthHandler(h.handleGetFamily)).Methods("GET")
    
    router.HandleFunc("/families/{id}/invites", h.authMiddleware.AuthHandler(h.handleCreateInvite)).Methods("POST")
    router.HandleFunc("/families/invites/{token}", h.handleValidateInvite).Methods("GET")
    
    router.HandleFunc("/families/{id}/modules", h.authMiddleware.AuthHandler(h.handleGetModules)).Methods("GET")
    router.HandleFunc("/families/{id}/modules/{moduleId}", h.authMiddleware.AuthHandler(h.handleUpdateModule)).Methods("PUT")
}

func (h *Handler) handleCreateFamily(w http.ResponseWriter, r *http.Request) {
    var req CreateFamilyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    userCtx := r.Context().Value("user").(*models.UserContext)
    req.OwnerID = userCtx.UserID

    family, err := h.service.CreateFamily(&req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusCreated, family)
}

func (h *Handler) handleGetFamily(w http.ResponseWriter, r *http.Request) {
    id, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    family, err := h.service.GetFamily(id)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, family)
}

func (h *Handler) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
    familyID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    var req CreateInviteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    req.FamilyID = familyID

    userCtx := r.Context().Value("user").(*models.UserContext)
    if userCtx.Role != models.RoleParent {
        writeError(w, http.StatusForbidden, "only parents can create invites")
        return
    }

    invite, err := h.service.CreateInvite(&req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusCreated, invite)
}

func (h *Handler) handleValidateInvite(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    token := vars["token"]

    invite, err := h.service.ValidateInvite(token)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, invite)
}

func (h *Handler) handleGetModules(w http.ResponseWriter, r *http.Request) {
    familyID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    modules, err := h.service.GetFamilyModules(familyID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, modules)
}

func (h *Handler) handleUpdateModule(w http.ResponseWriter, r *http.Request) {
    familyID, err := getIDFromRequest(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
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

    userCtx := r.Context().Value("user").(*models.UserContext)
    if userCtx.Role != models.RoleParent {
        writeError(w, http.StatusForbidden, "only parents can update modules")
        return
    }

    if err := h.service.UpdateModuleSettings(familyID, &req); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{"message": "module updated successfully"})
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