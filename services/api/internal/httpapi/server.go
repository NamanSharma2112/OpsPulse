package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/auth"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/config"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/ingest"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/orgs"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/projects"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

// Deps are everything the HTTP layer needs to serve requests.
type Deps struct {
	Config       config.Config
	Logger       *slog.Logger
	Auth         *auth.Service
	GitHubAuth   *auth.GitHubService
	Orgs         *orgs.Service
	Projects     *projects.Service
	Ingest       *ingest.Service
	Events       store.Events
	Metrics      store.Metrics
	Deployments  store.Deployments
	PullRequests store.PullRequests
	Incidents    store.Incidents
	Pinger       interface {
		Ping(context.Context) error
	}
}

// Server owns the HTTP router and its dependencies.
type Server struct {
	cfg          config.Config
	log          *slog.Logger
	auth         *auth.Service
	githubAuth   *auth.GitHubService
	orgs         *orgs.Service
	projects     *projects.Service
	ingest       *ingest.Service
	events       store.Events
	metrics      store.Metrics
	deployments  store.Deployments
	pullRequests store.PullRequests
	incidents    store.Incidents
	pinger       interface {
		Ping(context.Context) error
	}
	handler http.Handler
}

// New builds the server and its routes.
func New(d Deps) *Server {
	s := &Server{
		cfg:          d.Config,
		log:          d.Logger,
		auth:         d.Auth,
		githubAuth:   d.GitHubAuth,
		orgs:         d.Orgs,
		projects:     d.Projects,
		ingest:       d.Ingest,
		events:       d.Events,
		metrics:      d.Metrics,
		deployments:  d.Deployments,
		pullRequests: d.PullRequests,
		incidents:    d.Incidents,
		pinger:       d.Pinger,
	}
	s.handler = chain(s.routes(),
		recoverPanics(d.Logger),
		requestLogger(d.Logger),
		cors(d.Config.CORSOrigins),
	)
	return s
}

// ServeHTTP makes the server an http.Handler, which keeps it testable without
// binding a port.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.handler.ServeHTTP(w, r) }

// HTTPServer returns a configured net/http server with sane timeouts.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           s,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
