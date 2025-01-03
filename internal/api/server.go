package api

import (
	"log"
	"net/http"

	"github.com/chrisabs/storage/internal/config"
	"github.com/chrisabs/storage/internal/container"
	"github.com/chrisabs/storage/internal/item"
	"github.com/chrisabs/storage/internal/middleware"
	"github.com/chrisabs/storage/internal/platform/database"
	"github.com/chrisabs/storage/internal/search"
	"github.com/chrisabs/storage/internal/tag"
	"github.com/chrisabs/storage/internal/user"
	"github.com/gorilla/mux"
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

	authMiddleware := middleware.NewAuthMiddleware(s.config.JWTSecret)

	userRepo := user.NewRepository(s.db.DB)
	containerRepo := container.NewRepository(s.db.DB)
	itemRepo := item.NewRepository(s.db.DB)
	tagRepo := tag.NewRepository(s.db.DB)
	searchRepo := search.NewRepository(s.db.DB)

	userService := user.NewService(userRepo, s.config.JWTSecret)
	containerService := container.NewService(containerRepo)
	itemService := item.NewService(itemRepo)
	tagService := tag.NewService(tagRepo)
	searchService := search.NewService(searchRepo)

	userHandler := user.NewHandler(userService, authMiddleware)
	containerHandler := container.NewHandler(containerService, authMiddleware)
	itemHandler := item.NewHandler(
		itemService,
		containerService,
		authMiddleware,
	)
	tagHandler := tag.NewHandler(tagService, authMiddleware)
	searchHandler := search.NewHandler(searchService, authMiddleware)

	userHandler.RegisterRoutes(router)
	containerHandler.RegisterRoutes(router)
	itemHandler.RegisterRoutes(router)
	tagHandler.RegisterRoutes(router)
	searchHandler.RegisterRoutes(router)

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}
