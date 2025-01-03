package api

import (
	"log"
	"net/http"

	"github.com/chrisabs/storage/internal/container"
	"github.com/chrisabs/storage/internal/item"
	"github.com/chrisabs/storage/internal/platform/database"
	"github.com/chrisabs/storage/internal/tag"
	"github.com/chrisabs/storage/internal/user"
	"github.com/gorilla/mux"
)

type Server struct {
	listenAddr string
	db         *database.PostgresDB
}

func NewServer(listenAddr string, db *database.PostgresDB) *Server {
	return &Server{
		listenAddr: listenAddr,
		db:         db,
	}
}

func (s *Server) Run() {
	router := mux.NewRouter()

	userRepo := user.NewRepository(s.db.DB)
	containerRepo := container.NewRepository(s.db.DB)
	itemRepo := item.NewRepository(s.db.DB)
	tagRepo := tag.NewRepository(s.db.DB)

	userService := user.NewService(userRepo)
	containerService := container.NewService(containerRepo)
	itemService := item.NewService(itemRepo)
	tagService := tag.NewService(tagRepo)

	userHandler := user.NewHandler(userService)
	containerHandler := container.NewHandler(containerService)
	itemHandler := item.NewHandler(itemService, containerService)
	tagHandler := tag.NewHandler(tagService)

	userHandler.RegisterRoutes(router)
	containerHandler.RegisterRoutes(router)
	itemHandler.RegisterRoutes(router)
	tagHandler.RegisterRoutes(router)

	log.Println("server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}
