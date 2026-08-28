package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/adapters/http/middleware"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

type mockAuthenticator struct {
	getUserFn func(ctx context.Context, apiKey string) (*user.User, error)
}

func (m *mockAuthenticator) GetUserByAPIKey(ctx context.Context, apiKey string) (*user.User, error) {
	return m.getUserFn(ctx, apiKey)
}

func TestRequireAuth(t *testing.T) {
	activeUser := &user.User{
		ID:     uuid.New(),
		Status: string(user.Active),
	}

	tests := []struct {
		name                string
		authHeader          string
		setupAuth           func() *mockAuthenticator
		expectedStatus      int
		expectUserInContext bool
	}{
		{
			name:       "missing auth header",
			authHeader: "",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid format auth header",
			authHeader: "Basic invalidformat",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid api key",
			authHeader: "Bearer bad_key",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{
					getUserFn: func(ctx context.Context, apiKey string) (*user.User, error) {
						return nil, user.ErrInvalidAPIKey
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "user suspended",
			authHeader: "Bearer suspended_key",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{
					getUserFn: func(ctx context.Context, apiKey string) (*user.User, error) {
						return nil, user.ErrUserSuspended
					},
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "user deleted",
			authHeader: "Bearer deleted_key",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{
					getUserFn: func(ctx context.Context, apiKey string) (*user.User, error) {
						return nil, user.ErrUserDeleted
					},
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "success active user",
			authHeader: "Bearer active_key",
			setupAuth: func() *mockAuthenticator {
				return &mockAuthenticator{
					getUserFn: func(ctx context.Context, apiKey string) (*user.User, error) {
						return activeUser, nil
					},
				}
			},
			expectedStatus:      http.StatusOK,
			expectUserInContext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := tt.setupAuth()
			mw := middleware.RequireAuth(auth)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectUserInContext {
					u := middleware.UserFromContext(r.Context())
					if u == nil || u.ID != activeUser.ID {
						t.Error("expected user to be in request context")
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			mw(nextHandler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
