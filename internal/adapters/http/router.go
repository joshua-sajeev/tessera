// Package http provides the HTTP adapter for Tessera.
package http

import (
	"net/http"

	"github.com/joshua-sajeev/tessera/internal/adapters/http/handler"
	"github.com/joshua-sajeev/tessera/internal/adapters/http/middleware"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// NewRouter constructs a new HTTP router with registered routes.
func NewRouter(userHandler *handler.UserHandler, authenticator ports.Authenticator) http.Handler {
	mux := http.NewServeMux()

	authMiddleware := middleware.RequireAuth(authenticator)

	// Register routes
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.Handle("GET /users/{id}", authMiddleware(http.HandlerFunc(userHandler.Get)))
	mux.Handle("PUT /users/{id}/status", authMiddleware(http.HandlerFunc(userHandler.UpdateStatus)))
	mux.Handle("PATCH /users/{id}/status", authMiddleware(http.HandlerFunc(userHandler.UpdateStatus)))

	return mux
}
