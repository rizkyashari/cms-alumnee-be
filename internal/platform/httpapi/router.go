package httpapi

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/rizkyashari/cms-alumnee-be/internal/modules/auth"
	"github.com/rizkyashari/cms-alumnee-be/internal/modules/health"
	"github.com/rizkyashari/cms-alumnee-be/middleware"
)

func NewRouter(logger *slog.Logger, db health.ReadinessChecker) *chi.Mux {
	return NewRouterWithAuth(logger, db, nil)
}

func NewRouterWithAuth(logger *slog.Logger, db health.ReadinessChecker, authHandler *auth.Handler, modules ...func(chi.Router)) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.SafeRecovery(logger))
	health.RegisterRoutes(router, db)
	if authHandler != nil {
		authHandler.RegisterRoutes(router)
	}
	for _, register := range modules {
		register(router)
	}
	return router
}
