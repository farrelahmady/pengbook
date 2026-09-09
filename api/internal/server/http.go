// Package server wires the HTTP layer: chi router, middleware, and graceful
// shutdown.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger"

	inner "pengbook/api/internal/middleware"
	"pengbook/api/internal/module/account"
	"pengbook/api/internal/module/auth"
	"pengbook/api/internal/module/journal"
	"pengbook/api/internal/module/user"
	"pengbook/api/pkg/response"

	_ "pengbook/api/docs"
)

// HTTP wraps the http.Server for the application.
type HTTP struct {
	srv *http.Server
}

// NewHTTP builds the chi router with middleware and all mounted routes.
//
// Middleware chain: RequestID → RealIP → Recovery → Logger → Timeout.
// Module routers are mounted under their versioned paths (e.g. /api/v1/users),
// plus a /health endpoint for liveness checks.
func NewHTTP(userHandler *user.Handler, authHandler *auth.Handler, accountHandler *account.Handler, journalHandler *journal.Handler, jwtSecret string, port int, log *slog.Logger) *HTTP {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(inner.CORS)
	r.Use(inner.Recovery(log))
	r.Use(inner.Logger(log))
	r.Use(middleware.Timeout(15 * time.Second))

	r.Route("/api/v1/users", func(r chi.Router) {
		r.Mount("/", userHandler.Routes())
	})

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Mount("/", authHandler.Routes())
	})

	// Protected routes (require JWT)
	r.Route("/api/v1/accounts", func(r chi.Router) {
		r.Use(inner.Auth(jwtSecret))
		r.Mount("/", accountHandler.Routes())
	})

	r.Route("/api/v1/journals", func(r chi.Router) {
		r.Use(inner.Auth(jwtSecret))
		r.Mount("/", journalHandler.Routes())
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler())

	return &HTTP{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
	}
}

// Run starts the HTTP server and blocks until it is shut down.
//
// It listens for SIGINT/SIGTERM; on signal it performs a graceful shutdown
// with a 10s timeout. Real listen errors are returned; ErrServerClosed
// (normal shutdown) is ignored.
func (h *HTTP) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return h.srv.Shutdown(shutdownCtx)
	}
}
