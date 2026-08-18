package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"gcet-web-backend/internal/logger"
	"gcet-web-backend/internal/models"
	"gcet-web-backend/internal/service"
)

type contextKey string

const (
	ClaimsKey  contextKey = "claims"
	SessionKey contextKey = "session"
)

// AuthMiddleware validates JWT tokens and injects claims into context
func AuthMiddleware(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				writeJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "Authorization token required",
				})
				return
			}

			claims, err := authSvc.ValidateToken(tokenStr)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "Invalid or expired token",
				})
				return
			}

			// Get the live session
			session, err := authSvc.GetSession(claims.SessionID)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "Session expired — please login again",
				})
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			ctx = context.WithValue(ctx, SessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves JWT claims from context
func GetClaims(r *http.Request) *models.TokenClaims {
	if claims, ok := r.Context().Value(ClaimsKey).(*models.TokenClaims); ok {
		return claims
	}
	return nil
}

// GetSession retrieves the faculty session from context
func GetSession(r *http.Request) *service.FacultySession {
	if sess, ok := r.Context().Value(SessionKey).(*service.FacultySession); ok {
		return sess
	}
	return nil
}

// RequestLogger logs HTTP requests
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)

		logger.Log.Infof("%s %s %d %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.String(),
		)
	})
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// RequestTimeout sets a timeout on requests
func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}
	// Check cookie
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}
	return ""
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
