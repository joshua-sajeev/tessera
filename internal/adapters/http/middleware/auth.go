package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

type contextKey string

const userContextKey contextKey = "user"

// WithUser returns context with authenticated user.
func WithUser(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserFromContext gets user from context.
func UserFromContext(ctx context.Context) *user.User {
	u, _ := ctx.Value(userContextKey).(*user.User)
	return u
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// RequireAuth guards endpoints.
func RequireAuth(authenticator ports.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				respondError(w, http.StatusUnauthorized, "invalid authorization header format; must be Bearer <token>")
				return
			}

			apiKey := parts[1]
			u, err := authenticator.GetUserByAPIKey(r.Context(), apiKey)
			if err != nil {
				if errors.Is(err, user.ErrInvalidAPIKey) || errors.Is(err, user.ErrUserNotFound) {
					respondError(w, http.StatusUnauthorized, "invalid API key")
					return
				}
				if errors.Is(err, user.ErrUserSuspended) {
					respondError(w, http.StatusForbidden, "user is suspended")
					return
				}
				if errors.Is(err, user.ErrUserDeleted) {
					respondError(w, http.StatusForbidden, "user is deleted")
					return
				}
				respondError(w, http.StatusInternalServerError, "authentication failed")
				return
			}

			ctx := WithUser(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
