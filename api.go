package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

type APIServer struct {
	addr   string
	db     *sql.DB
	router *chi.Mux
}

func NewAPIServer(addr string) *APIServer {

	newServer := &APIServer{
		addr:   addr,
		router: chi.NewRouter(),
	}

	return newServer
}

func initRoutes() error {

	apiRouter := chi.NewRouter()

	server.router.Mount("/api/v1", apiRouter)

	apiRouter.Get("/users", getUsersHandler)
	apiRouter.Post("/register", registerHandler)
	apiRouter.Post("/login", loginHandler)

	apiRouter.Get("/users/{id}", getUserHandler)

	return nil
}
