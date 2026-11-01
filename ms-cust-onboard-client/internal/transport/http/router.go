package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"ms-cust-onboard-client/internal/observability"
)

// SetupRouter sets up Chi routing for ms-cust-onboard-client
func SetupRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(observability.TracingMiddleware("ms-cust-onboard-client"))
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Trace-Id", "X-Span-Id", "X-Request-Id", "traceparent"},
		ExposedHeaders:   []string{"Link", "X-Trace-Id", "X-Span-Id", "traceparent"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", h.HealthCheckHandler)
	r.Get("/health", h.HealthCheckHandler)

	r.Route("/api/v1/client/onboarding", func(r chi.Router) {
		r.Post("/start", h.StartSessionHandler)
		r.Post("/resume", h.ResumeSessionHandler)
		r.Post("/upload", h.UploadDocumentHandler)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetApplicationHandler)
			r.Post("/cancel", h.CancelSessionHandler)
			r.Post("/submit", h.SubmitApplicationHandler)

			// Step patches
			r.Patch("/personal-info", h.PatchPersonalInfoHandler)
			r.Patch("/address", h.PatchAddressHandler)
			r.Patch("/identity-documents", h.PatchIdentityDocsHandler)
			r.Patch("/financial-info", h.PatchFinancialInfoHandler)
			r.Patch("/wallet-tier", h.PatchWalletTierHandler)
			r.Patch("/draft", h.PatchDraftHandler)

			// Resubmit, audit & progress
			r.Post("/resubmit-documents", h.ResubmitDocumentsHandler)
			r.Get("/audit-logs", h.GetAuditLogsHandler)
			r.Get("/workflow-progress", h.GetWorkflowProgressHandler)
		})
	})

	return r
}
