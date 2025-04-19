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

func initRoutes() {

	apiRouter := chi.NewRouter()

	// --- Public API Routes ---
	apiRouter.Post("/register", registerHandler)
	apiRouter.Post("/login", logInHandler)

	// --- Protected API Routes ---
	apiRouter.Group(func(protectedRouter chi.Router) {
		protectedRouter.Use(AuthenticationMiddleware, CSRFProtectionMiddleware)

		protectedRouter.Get("/users", getUsersHandler)
		protectedRouter.Post("/logout", logOutHandler)
		protectedRouter.Get("/users/{username}", getUserHandler)
		protectedRouter.Get("/protected", protectedHandler)
		protectedRouter.Post("/protected", protectedHandler)

	})

	server.router.Use(LoggingMiddleware, RateLimiterMiddleware)

	server.router.Mount("/api/v1", apiRouter)
}
