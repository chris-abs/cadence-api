package api

import (
	"log"
	"net/http"

	"github.com/chrisabs/storage/internal/container"
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

	containerRepo := container.NewRepository(s.db.DB)
	containerService := container.NewService(containerRepo)
	containerHandler := container.NewHandler(containerService)
	containerHandler.RegisterRoutes(router)

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}
