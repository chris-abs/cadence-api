package user

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
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
   // Public routes
   router.HandleFunc("/users/register", h.handleRegister).Methods("POST")
   router.HandleFunc("/users/login", h.handleLogin).Methods("POST")
   router.HandleFunc("/users/accept-invite", h.handleAcceptInvite).Methods("POST")

   // Protected routes
   router.HandleFunc("/users", h.authMiddleware.AuthHandler(h.handleGetUsers)).Methods("GET")
   router.HandleFunc("/user", h.authMiddleware.AuthHandler(h.handleGetAuthenticatedUser)).Methods("GET")
   router.HandleFunc("/users/{id}", h.authMiddleware.AuthHandler(h.handleGetUser)).Methods("GET")
   router.HandleFunc("/users/{id}", h.authMiddleware.AuthHandler(h.handleUpdateUser)).Methods("PUT")
   router.HandleFunc("/users/{id}", h.authMiddleware.AuthHandler(h.handleDeleteUser)).Methods("DELETE")
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
   var req CreateUserRequest
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
       writeError(w, http.StatusBadRequest, "invalid request body")
       return
   }

   user, err := h.service.CreateUser(&req)
   if err != nil {
       writeError(w, http.StatusInternalServerError, err.Error())
       return
   }

   token, err := h.service.generateJWT(user.ID)
   if err != nil {
       writeError(w, http.StatusInternalServerError, "failed to generate token")
       return
   }

   writeJSON(w, http.StatusCreated, &AuthResponse{
       Token: token,
       User:  *user,
   })
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

func (h *Handler) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
   var req AcceptInviteRequest
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
       writeError(w, http.StatusBadRequest, "invalid request body")
       return
   }

   user, err := h.service.AcceptInvite(&req)
   if err != nil {
       writeError(w, http.StatusBadRequest, err.Error())
       return
   }

   token, err := h.service.generateJWT(user.ID)
   if err != nil {
       writeError(w, http.StatusInternalServerError, "failed to generate token")
       return
   }

   writeJSON(w, http.StatusOK, &AuthResponse{
       Token: token,
       User:  *user,
   })
}

func (h *Handler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
   userCtx := r.Context().Value("user").(*models.UserContext)
   
   if userCtx.Role != models.RoleParent {
       writeError(w, http.StatusForbidden, "only parents can view all users")
       return
   }

   users, err := h.service.GetAllUsers()
   if err != nil {
       writeError(w, http.StatusInternalServerError, err.Error())
       return
   }

   writeJSON(w, http.StatusOK, users)
}

func (h *Handler) handleGetAuthenticatedUser(w http.ResponseWriter, r *http.Request) {
   userCtx := r.Context().Value("user").(*models.UserContext)
   
   user, err := h.service.GetUserByID(userCtx.UserID)
   if err != nil {
       writeError(w, http.StatusNotFound, err.Error())
       return
   }

   writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
   userCtx := r.Context().Value("user").(*models.UserContext)
   
   id, err := getIDFromRequest(r)
   if err != nil {
       writeError(w, http.StatusBadRequest, err.Error())
       return
   }

   user, err := h.service.GetUserByID(id)
   if err != nil {
       writeError(w, http.StatusNotFound, err.Error())
       return
   }

   if user.FamilyID != userCtx.FamilyID {
       writeError(w, http.StatusForbidden, "access denied")
       return
   }

   writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
   userCtx := r.Context().Value("user").(*models.UserContext)
   
   id, err := getIDFromRequest(r)
   if err != nil {
       writeError(w, http.StatusBadRequest, err.Error())
       return
   }

   if id != userCtx.UserID && userCtx.Role != models.RoleParent {
       writeError(w, http.StatusForbidden, "access denied")
       return
   }

   if err := r.ParseMultipartForm(10 << 20); err != nil {
       writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse form: %v", err))
       return
   }

   firstName := r.FormValue("firstName")
   if firstName == "" {
       writeError(w, http.StatusBadRequest, "firstName is required")
       return
   }

   lastName := r.FormValue("lastName")
   if lastName == "" {
       writeError(w, http.StatusBadRequest, "lastName is required")
       return
   }

   var imageFile *multipart.FileHeader
   if file, header, err := r.FormFile("image"); err == nil {
       defer file.Close()
       imageFile = header
   }

   user, err := h.service.UpdateUser(id, firstName, lastName, imageFile)
   if err != nil {
       writeError(w, http.StatusInternalServerError, err.Error())
       return
   }

   writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
   userCtx := r.Context().Value("user").(*models.UserContext)
   
   id, err := getIDFromRequest(r)
   if err != nil {
       writeError(w, http.StatusBadRequest, err.Error())
       return
   }

   if userCtx.Role != models.RoleParent {
       writeError(w, http.StatusForbidden, "only parents can delete users")
       return
   }

   if err := h.service.DeleteUser(id); err != nil {
       writeError(w, http.StatusInternalServerError, err.Error())
       return
   }

   writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted successfully"})
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