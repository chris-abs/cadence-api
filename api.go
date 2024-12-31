package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type APISERVER struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APISERVER {
	return &APISERVER{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APISERVER) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/containers", makeHTTPHandleFunc(s.handleGetContainers)).Methods("GET")
	router.HandleFunc("/containers", makeHTTPHandleFunc(s.handleCreateContainer)).Methods("POST")
	router.HandleFunc("/containers/{id}", makeHTTPHandleFunc(s.handleGetContainerById)).Methods("GET")
	router.HandleFunc("/containers/{id}", makeHTTPHandleFunc(s.handleDeleteContainer)).Methods("DELETE")
	router.HandleFunc("/containers/{id}", makeHTTPHandleFunc(s.handleUpdateContainer)).Methods("PUT")

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APISERVER) handleGetContainers(w http.ResponseWriter, r *http.Request) error {
	containers, err := s.store.GetContainers()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, containers)
}

func (s *APISERVER) handleGetContainerById(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid id provided %s", idStr)
	}
	fmt.Println(id)

	container, err := s.store.GetContainerByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, container)
}

func (s *APISERVER) handleCreateContainer(w http.ResponseWriter, r *http.Request) error {
	createContainerRequest := new(CreateContainerRequest)
	if err := json.NewDecoder(r.Body).Decode(createContainerRequest); err != nil {
		return fmt.Errorf("invalid request body: %v", err)
	}

	container := NewContainer(createContainerRequest.Name, createContainerRequest.Location)
	if err := s.store.CreateContainer(container); err != nil {
		return fmt.Errorf("error creating container: %v", err)
	}

	return WriteJSON(w, http.StatusCreated, container)
}

func (s *APISERVER) handleUpdateContainer(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Implement update logic
	fmt.Printf("Update container: %s\n", id)
	return nil
}

func (s *APISERVER) handleDeleteContainer(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id := vars["id"]

	// TODO: Implement delete logic
	fmt.Printf("Delete container: %s\n", id)
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
