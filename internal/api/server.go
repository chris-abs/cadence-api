package api

import (
	"log"
	"net/http"

	"github.com/chrisabs/cadence/internal/chores"
	"github.com/chrisabs/cadence/internal/config"
	"github.com/chrisabs/cadence/internal/family"
	"github.com/chrisabs/cadence/internal/membership"
	"github.com/chrisabs/cadence/internal/middleware"
	"github.com/chrisabs/cadence/internal/platform/database"
	"github.com/chrisabs/cadence/internal/profile"
	"github.com/chrisabs/cadence/internal/storage/container"
	"github.com/chrisabs/cadence/internal/storage/item"
	"github.com/chrisabs/cadence/internal/storage/recent"
	"github.com/chrisabs/cadence/internal/storage/search"
	"github.com/chrisabs/cadence/internal/storage/tag"
	"github.com/chrisabs/cadence/internal/storage/workspace"
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
	profileRepo := profile.NewRepository(s.db.DB)
	familyRepo := family.NewRepository(s.db.DB)
	membershipRepo := membership.NewRepository(s.db.DB)
	containerRepo := container.NewRepository(s.db.DB)
	workspaceRepo := workspace.NewRepository(s.db.DB)
	itemRepo := item.NewRepository(s.db.DB)
	tagRepo := tag.NewRepository(s.db.DB)
	searchRepo := search.NewRepository(s.db.DB)
	recentRepo := recent.NewRepository(s.db.DB)
	choreRepo := chores.NewRepository(s.db.DB)  

	profileService := profile.NewService(
		profileRepo,
		nil, 
		s.config.JWTSecret,
	)
	
	membershipService := membership.NewService(membershipRepo)
	
	familyService := family.NewService(
		familyRepo,
		profileService,
		membershipService,
	)
	
	profileService.SetMembershipService(membershipService)
	
	// Initialize other services
	workspaceService := workspace.NewService(workspaceRepo)
	containerService := container.NewService(containerRepo)
	itemService := item.NewService(itemRepo)
	tagService := tag.NewService(tagRepo)
	searchService := search.NewService(searchRepo)
	recentService := recent.NewService(recentRepo)
	choreService := chores.NewService(choreRepo) 

	// Initialise auth middleware with profile validation
	authMiddleware := middleware.NewAuthMiddleware(
		s.config.JWTSecret,
		s.db.DB,
		membershipService,
		familyService, 
	)

	// Initialise handlers
	profileHandler := profile.NewHandler(profileService, authMiddleware, membershipService)
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
	choreHandler := chores.NewHandler(choreService, authMiddleware)  

	// Register routes
	profileHandler.RegisterRoutes(router)
	familyHandler.RegisterRoutes(router)
	membershipHandler.RegisterRoutes(router)
	workspaceHandler.RegisterRoutes(router)
	containerHandler.RegisterRoutes(router)
	itemHandler.RegisterRoutes(router)
	tagHandler.RegisterRoutes(router)
	searchHandler.RegisterRoutes(router)
	recentHandler.RegisterRoutes(router)
	choreHandler.RegisterRoutes(router)  

	handler := c.Handler(router)

	log.Printf("JSON API server running on port: %s", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}