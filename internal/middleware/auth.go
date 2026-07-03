package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"error":   "unauthorized",
					"message": "No bearer",
				})
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"error":   "unauthorized",
					"message": "Token not valid",
				})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"error":   "unauthorized",
					"message": "No claims",
				})
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"error":   "unauthorized",
					"message": "No user_id",
				})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
