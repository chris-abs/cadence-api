package api

import (
	"log"
	"net/http"

	"github.com/chrisabs/storage/internal/container"
	"github.com/chrisabs/storage/internal/item"
	"github.com/chrisabs/storage/internal/platform/database"
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

	// Container setup
	containerRepo := container.NewRepository(s.db.DB)
	containerService := container.NewService(containerRepo)
	containerHandler := container.NewHandler(containerService)
	containerHandler.RegisterRoutes(router)

	// Item setup
	itemRepo := item.NewRepository(s.db.DB)
	itemService := item.NewService(itemRepo)
	itemHandler := item.NewHandler(itemService)
	itemHandler.RegisterRoutes(router)

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}
