package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/albnnaardy11/pahlawan-pangan/internal/auth/domain"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(uc domain.AuthUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}

			user, err := uc.ValidateToken(r.Context(), parts[1])
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Inject user into context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleGuard ensures only specific roles can access a route
func RoleGuard(roles ...domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*domain.User)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			allowed := false
			for _, role := range roles {
				if user.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
