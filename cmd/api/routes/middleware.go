package routes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"golang.org/x/time/rate"
)

var apiKeys = map[string]*rate.Limiter{
	"9686032971eb850a002a86649e468253cc76e1e2": rate.NewLimiter(1, 2),  // 1 request per second, burst size of 2
	"24af1b0df83bc86621ce305203d9d11b5f698e62": rate.NewLimiter(5, 10), // 5 requests per second, burst size of 10
	"1edae2c22e997a5a3f1609e8832cea285abebcde": rate.NewLimiter(10000, 1000),
}

func GenerateAPIKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type limiterKeyType string

const limiterKey = limiterKeyType("limiter")

func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		limiter, exists := apiKeys[apiKey]
		if !exists {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), limiterKey, limiter)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter, ok := r.Context().Value(limiterKey).(*rate.Limiter)

		if !ok || !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
