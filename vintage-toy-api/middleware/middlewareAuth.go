package middleware

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware validates JWT from cookies
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Middleware: Received Cookies:", r.Cookies())

		cookie, err := r.Cookie("auth_token")
		if err != nil {
			log.Println("Middleware: No auth_token cookie found")
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		log.Println("Middleware: Extracted Token:", tokenStr)

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			log.Println("❌ Middleware: Invalid or expired token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			log.Println("❌ Middleware: Invalid token subject")
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
