package httpapi

import (
	"log/slog"

	"github.com/rizkyashari/cms-alumnee-be/internal/modules/health"
	"github.com/rizkyashari/cms-alumnee-be/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(logger *slog.Logger, db health.ReadinessChecker) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.SafeRecovery(logger))
	health.RegisterRoutes(router, db)
	return router
}
