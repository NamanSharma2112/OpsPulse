// Command api runs the OpsPulse HTTP service: it authenticates users, manages
// projects, ingests GitHub webhooks and serves the dashboard's reads.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/auth"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/config"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/httpapi"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/ingest"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/observability"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/orgs"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/projects"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store/postgres"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := observability.NewLogger(cfg.LogLevel)
	slog.SetDefault(log)

	// Signals cancel the root context, which unwinds the whole service.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	log.Info("connected to database")

	tokens := auth.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)
	authSvc := auth.NewService(db.Users, tokens)
	orgSvc := orgs.NewService(db.Orgs)
	projectSvc := projects.NewService(db.Projects, orgSvc)
	ingestSvc := ingest.NewService(db.Events, db.Metrics, db.Deployments, db.PullRequests, db.Incidents, log)

	server := httpapi.New(httpapi.Deps{
		Config:       cfg,
		Logger:       log,
		Auth:         authSvc,
		Orgs:         orgSvc,
		Projects:     projectSvc,
		Ingest:       ingestSvc,
		Events:       db.Events,
		Metrics:      db.Metrics,
		Deployments:  db.Deployments,
		PullRequests: db.PullRequests,
		Incidents:    db.Incidents,
		Pinger:       db,
	})

	httpServer := server.HTTPServer()
	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	// Give in-flight requests a bounded window to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("shutdown complete")
	return nil
}
