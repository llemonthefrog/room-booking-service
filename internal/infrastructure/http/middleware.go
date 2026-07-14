package http

import (
	"avito-task/internal/contracts/usecase"
	"avito-task/internal/domain"
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Middleware func(handler http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

func AuthMiddleware(verifier usecase.JWTUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				renderError(
					w,
					http.StatusUnauthorized,
					ErrCodeUnauthorized,
					"missing authorization header",
				)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				renderError(
					w,
					http.StatusUnauthorized,
					ErrCodeUnauthorized,
					"invalid authorization header format",
				)
				return
			}

			tokenString := parts[1]

			claims, err := verifier.VerifyToken(tokenString)
			if err != nil {
				renderError(
					w,
					http.StatusUnauthorized,
					ErrCodeUnauthorized,
					"invalid or expired token",
				)
				return
			}

			parsedUid, err := uuid.Parse(claims.UserID)
			if err != nil {
				renderError(
					w,
					http.StatusUnauthorized,
					ErrCodeUnauthorized,
					"invalid user id format in token",
				)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", parsedUid)
			ctx = context.WithValue(ctx, "role", string(claims.Role))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(requiredRole domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("role").(string)

			if !ok || role != string(requiredRole) {
				w.Header().Set("Content-Type", "application/json")
				renderError(w, http.StatusForbidden, ErrCodeForbidden, "access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
