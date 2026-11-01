package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"ms-eligiblity/internal/observability"
)

// SetupRouter configures Chi routes, middlewares and handlers
func SetupRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(observability.TracingMiddleware("ms-eligiblity"))
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Trace-Id", "X-Span-Id", "X-Request-Id", "traceparent"},
		ExposedHeaders:   []string{"Link", "X-Trace-Id", "X-Span-Id", "traceparent"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health and readiness
	r.Get("/healthz", h.HealthCheckHandler)
	r.Get("/health", h.HealthCheckHandler)

	// API v1 routes
	r.Route("/api/v1/eligibility", func(r chi.Router) {
		r.Post("/check", h.CheckEligibilityHandler)
		r.Get("/rules", h.ListRulesHandler)
		r.Get("/checks/{id}", h.GetCheckByIDHandler)
		r.Get("/customers/{customerId}/latest", h.GetLatestByCustomerIDHandler)
	})

	return r
}
