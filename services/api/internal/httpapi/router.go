package httpapi

import "net/http"

// routes registers every endpoint. Patterns use the method-aware routing that
// net/http gained in Go 1.22, so no third-party router is needed.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Liveness and readiness, used by Docker and load balancers.
	mux.HandleFunc("GET /healthz", s.handleLiveness)
	mux.HandleFunc("GET /readyz", s.handleReadiness)

	// Auth.
	mux.HandleFunc("POST /v1/auth/register", s.handleRegister)
	mux.HandleFunc("POST /v1/auth/login", s.handleLogin)
	mux.Handle("GET /v1/auth/me", s.requireAuth(http.HandlerFunc(s.handleMe)))
	mux.HandleFunc("GET /v1/auth/providers", s.handleAuthProviders)
	mux.HandleFunc("POST /v1/auth/signout", s.handleSignOut)

	// Sign in with GitHub. The callback redirects back to the dashboard with
	// an httpOnly session cookie.
	mux.HandleFunc("GET /v1/auth/github", s.handleGitHubAuthorize)
	mux.HandleFunc("GET /v1/auth/github/callback", s.handleGitHubCallback)

	// Repositories the signed-in user can administer on GitHub.
	mux.Handle("GET /v1/github/repositories", s.requireAuth(http.HandlerFunc(s.handleListGitHubRepositories)))

	// Organizations.
	mux.Handle("GET /v1/organizations", s.requireAuth(http.HandlerFunc(s.handleListOrgs)))
	mux.Handle("POST /v1/organizations", s.requireAuth(http.HandlerFunc(s.handleCreateOrg)))
	mux.Handle("GET /v1/organizations/{organizationID}", s.requireAuth(http.HandlerFunc(s.handleGetOrg)))
	mux.Handle("POST /v1/organizations/{organizationID}/members", s.requireAuth(http.HandlerFunc(s.handleAddOrgMember)))
	mux.Handle("GET /v1/organizations/{organizationID}/projects", s.requireAuth(http.HandlerFunc(s.handleListOrgProjects)))

	// Projects and the repositories they watch.
	mux.Handle("GET /v1/projects", s.requireAuth(http.HandlerFunc(s.handleListProjects)))
	mux.Handle("POST /v1/projects", s.requireAuth(http.HandlerFunc(s.handleCreateProject)))
	mux.Handle("GET /v1/projects/{projectID}", s.requireAuth(http.HandlerFunc(s.handleGetProject)))
	mux.Handle("GET /v1/projects/{projectID}/repositories", s.requireAuth(http.HandlerFunc(s.handleListRepositories)))
	mux.Handle("POST /v1/projects/{projectID}/repositories", s.requireAuth(http.HandlerFunc(s.handleConnectRepository)))

	// Dashboard reads.
	mux.Handle("GET /v1/projects/{projectID}/health", s.requireAuth(http.HandlerFunc(s.handleProjectHealth)))
	mux.Handle("GET /v1/projects/{projectID}/deployments", s.requireAuth(http.HandlerFunc(s.handleListDeployments)))
	mux.Handle("GET /v1/projects/{projectID}/pull-requests", s.requireAuth(http.HandlerFunc(s.handleListPullRequests)))
	mux.Handle("GET /v1/projects/{projectID}/incidents", s.requireAuth(http.HandlerFunc(s.handleListIncidents)))
	mux.Handle("GET /v1/projects/{projectID}/events", s.requireAuth(http.HandlerFunc(s.handleListEvents)))

	// GitHub webhook ingest. Authenticated by delivery signature, not a token.
	mux.HandleFunc("POST /v1/webhooks/github", s.handleGitHubWebhook)

	return mux
}
