package api

import (
	"time"

	"gcet-web-backend/internal/config"
	"gcet-web-backend/internal/middleware"
	"gcet-web-backend/internal/repository"
	"gcet-web-backend/internal/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes(r chi.Router, cfg *config.Config, authSvc *service.AuthService, repo *repository.SupabaseRepo) {
	h := NewHandler(authSvc, repo)

	// Global Middleware
	r.Use(middleware.RequestLogger)
	r.Use(middleware.SecurityHeaders)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting (10 req/sec by IP)
	r.Use(httprate.LimitByIP(cfg.RateLimitRPS, 1*time.Second))

	// API v1 router
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.HealthCheck)

		// Public Routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", h.Login)
			r.Post("/auth/refresh", h.RefreshToken)
		})

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authSvc))
			r.Use(middleware.RequestTimeout(30 * time.Second))

			// Auth
			r.Post("/auth/logout", h.Logout)

			// Profile
			r.Get("/profile", h.GetProfile)

			// Feedback
			r.Post("/feedback", h.SubmitFeedback)

			// Timetable
			r.Route("/timetable", func(r chi.Router) {
				r.Get("/", h.GetTimetable)
				r.Post("/custom", h.AddCustomTimetableEntry)
				r.Delete("/custom", h.DeleteCustomTimetableEntry)
				r.Post("/hide", h.HideScrapedTimetableEntry)
				r.Post("/reset", h.ResetTimetable)
			})

			// Attendance
			r.Route("/attendance", func(r chi.Router) {
				r.Get("/sheets", h.GetAttendanceSheets)
				r.Get("/courses", h.GetAttendanceCourses)
				r.Post("/students", h.GetStudentList)
				r.Post("/enter", h.SubmitAttendance)
				r.Get("/avg", h.GetAverageAttendance)
				r.Post("/student-wise", h.GetStudentWiseAttendance)
			})
		})
	})
}
