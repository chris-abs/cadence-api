package api

import (
	"log"
	"net/http"

	"github.com/chrisabs/storage/internal/config"
	"github.com/chrisabs/storage/internal/container"
	"github.com/chrisabs/storage/internal/family"
	"github.com/chrisabs/storage/internal/item"
	"github.com/chrisabs/storage/internal/membership"
	"github.com/chrisabs/storage/internal/middleware"
	"github.com/chrisabs/storage/internal/platform/database"
	"github.com/chrisabs/storage/internal/recent"
	"github.com/chrisabs/storage/internal/search"
	"github.com/chrisabs/storage/internal/tag"
	"github.com/chrisabs/storage/internal/user"
	"github.com/chrisabs/storage/internal/workspace"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	listenAddr string
	db         *database.PostgresDB
	config     *config.Config
}

func NewServer(listenAddr string, db *database.PostgresDB, config *config.Config) *Server {
	return &Server{
		listenAddr: listenAddr,
		db:         db,
		config:     config,
	}
}

func (s *Server) Run() {
	router := mux.NewRouter()

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Initialise repositories
	userRepo := user.NewRepository(s.db.DB)
	familyRepo := family.NewRepository(s.db.DB)
	membershipRepo := membership.NewRepository(s.db.DB)
	containerRepo := container.NewRepository(s.db.DB)
	workspaceRepo := workspace.NewRepository(s.db.DB)
	itemRepo := item.NewRepository(s.db.DB)
	tagRepo := tag.NewRepository(s.db.DB)
	searchRepo := search.NewRepository(s.db.DB)
	recentRepo := recent.NewRepository(s.db.DB)

	userService := user.NewService(
		userRepo,
		nil, 
		s.config.JWTSecret,
	)
	
	membershipService := membership.NewService(membershipRepo)
	
	familyService := family.NewService(
		familyRepo,
		userService,
		membershipService,
	)
	
	// Now we can set circular dependencies
	userService.SetMembershipService(membershipService)
	
	// Initialize other services
	workspaceService := workspace.NewService(workspaceRepo)
	containerService := container.NewService(containerRepo)
	itemService := item.NewService(itemRepo)
	tagService := tag.NewService(tagRepo)
	searchService := search.NewService(searchRepo)
	recentService := recent.NewService(recentRepo)

	// Initialise auth middleware with user validation
	authMiddleware := middleware.NewAuthMiddleware(
		s.config.JWTSecret,
		s.db.DB,
		membershipService,
		familyService, 
	)

	// Initialise handlers
	userHandler := user.NewHandler(userService, authMiddleware)
	familyHandler := family.NewHandler(
		familyService,
		authMiddleware,
	)
	membershipHandler := membership.NewHandler(membershipService, authMiddleware)
	workspaceHandler := workspace.NewHandler(workspaceService, authMiddleware)
	containerHandler := container.NewHandler(containerService, authMiddleware)
	itemHandler := item.NewHandler(itemService, containerService, authMiddleware)
	tagHandler := tag.NewHandler(tagService, authMiddleware)
	searchHandler := search.NewHandler(searchService, authMiddleware)
	recentHandler := recent.NewHandler(recentService, authMiddleware)

	// Register routes
	userHandler.RegisterRoutes(router)
	familyHandler.RegisterRoutes(router)
	membershipHandler.RegisterRoutes(router)
	workspaceHandler.RegisterRoutes(router)
	containerHandler.RegisterRoutes(router)
	itemHandler.RegisterRoutes(router)
	tagHandler.RegisterRoutes(router)
	searchHandler.RegisterRoutes(router)
	recentHandler.RegisterRoutes(router)

	handler := c.Handler(router)

	log.Printf("JSON API server running on port: %s", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}