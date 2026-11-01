package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"ms-customer-onboard/internal/observability"
)

// SetupRouter configures Chi routes, middlewares and onboarding endpoints
func SetupRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(observability.TracingMiddleware("ms-customer-onboard"))
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// CORS configuration for Frontend interaction
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Trace-Id", "X-Span-Id", "X-Request-Id", "traceparent"},
		ExposedHeaders:   []string{"Link", "X-Trace-Id", "X-Span-Id", "traceparent"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health probes
	r.Get("/healthz", h.HealthCheckHandler)
	r.Get("/health", h.HealthCheckHandler)

	// Onboarding API v1
	r.Route("/api/v1/onboarding", func(r chi.Router) {
		r.Post("/start", h.StartSessionHandler)
		r.Get("/applications", h.ListApplicationsHandler)
		r.Get("/phone/{phoneNumber}/active", h.GetActiveByPhoneNumberHandler)
		r.Get("/active", h.GetActiveByPhoneNumberHandler)
		r.Get("/customers/{customerId}/active", h.GetActiveByCustomerIDHandler)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetApplicationHandler)
			r.Post("/resume", h.ResumeSessionHandler)
			r.Post("/cancel", h.CancelSessionHandler)
			r.Post("/submit", h.SubmitApplicationHandler)

			// Granular step patches
			r.Patch("/personal-info", h.PatchPersonalInfoHandler)
			r.Patch("/address", h.PatchAddressHandler)
			r.Patch("/identity-documents", h.PatchIdentityDocsHandler)
			r.Patch("/financial-info", h.PatchFinancialInfoHandler)
			r.Patch("/wallet-tier", h.PatchWalletTierHandler)
			r.Patch("/draft", h.PatchDraftHandler)

			// Mid-flight review & document resolution
			r.Post("/resubmit-documents", h.ResubmitDocumentsHandler)
			r.Post("/manual-review", h.ManualReviewDecisionHandler)
			r.Put("/confirm-data", h.PutConfirmedDataHandler)
			r.Get("/audit-logs", h.GetAuditLogsHandler)
			r.Get("/workflow-progress", h.GetWorkflowProgressHandler)
		})
	})

	return r
}
