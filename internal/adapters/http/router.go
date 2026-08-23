// Package http provides the HTTP adapter for Tessera.
package http

import (
	"net/http"

	"github.com/joshua-sajeev/tessera/internal/adapters/http/handler"
	"github.com/joshua-sajeev/tessera/internal/application/userapp"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// NewRouter constructs a new HTTP router with registered routes.
func NewRouter(userRepo ports.UserRepository, apiKeyPrefix, apiKeyVersion string) http.Handler {
	mux := http.NewServeMux()

	userService := userapp.NewUserService(userRepo, apiKeyPrefix, apiKeyVersion)

	userHandler := handler.NewUserHandler(userService)

	// Register routes
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{id}", userHandler.Get)
	mux.HandleFunc("PUT /users/{id}/status", userHandler.UpdateStatus)
	mux.HandleFunc("PATCH /users/{id}/status", userHandler.UpdateStatus)

	return mux
}
