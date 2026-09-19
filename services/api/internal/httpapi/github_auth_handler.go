package httpapi

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	gh "github.com/NamanSharma2112/OpsPulse/services/api/internal/github"
)

// Cookie names. The state cookie lives only for the length of the round trip
// to GitHub; the session cookie is what the dashboard reads afterwards.
const (
	oauthStateCookie = "opspulse_oauth_state"
	sessionCookie    = "opspulse_session"
)

// handleGitHubAuthorize starts the OAuth flow.
//
// The state parameter is minted here and stored in a short-lived, httpOnly
// cookie. On the way back it must match, which is what stops a third party
// from completing a login into somebody else's session.
func (s *Server) handleGitHubAuthorize(w http.ResponseWriter, r *http.Request) {
	if !s.githubAuth.Configured() {
		writeError(w, http.StatusNotImplemented,
			"sign-in with GitHub is not configured on this server")
		return
	}

	state, err := gh.NewState()
	if err != nil {
		writeDomainError(w, s.log, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((10 * time.Minute).Seconds()),
	})
	http.Redirect(w, r, s.githubAuth.AuthorizeURL(state), http.StatusFound)
}

// handleGitHubCallback completes the flow and sends the browser back to the
// dashboard with a session cookie.
func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	if !s.githubAuth.Configured() {
		writeError(w, http.StatusNotImplemented,
			"sign-in with GitHub is not configured on this server")
		return
	}

	query := r.URL.Query()
	// GitHub reports a refused consent screen here rather than by status.
	if e := query.Get("error"); e != "" {
		s.redirectToApp(w, r, "/signin?error="+url.QueryEscape(firstNonEmpty(query.Get("error_description"), e)))
		return
	}

	state := query.Get("state")
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil || state == "" || cookie.Value != state {
		s.log.Warn("oauth state mismatch", "has_cookie", err == nil)
		s.redirectToApp(w, r, "/signin?error="+url.QueryEscape("sign-in expired, please try again"))
		return
	}
	// One round trip, one state.
	s.clearCookie(w, oauthStateCookie)

	code := query.Get("code")
	if code == "" {
		s.redirectToApp(w, r, "/signin?error="+url.QueryEscape("GitHub returned no authorisation code"))
		return
	}

	session, err := s.githubAuth.Complete(r.Context(), code)
	if err != nil {
		s.log.Error("github sign-in failed", "error", err)
		s.redirectToApp(w, r, "/signin?error="+url.QueryEscape("could not complete sign-in with GitHub"))
		return
	}

	// httpOnly: the dashboard reads this server-side and forwards it as a
	// bearer token, so script never needs to see it.
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})
	s.redirectToApp(w, r, "/dashboard")
}

// handleSignOut clears the session cookie.
func (s *Server) handleSignOut(w http.ResponseWriter, r *http.Request) {
	s.clearCookie(w, sessionCookie)
	writeJSON(w, http.StatusOK, map[string]any{"status": "signed out"})
}

// handleAuthProviders tells the dashboard which sign-in methods exist, so it
// can hide a GitHub button on a server that has no OAuth app configured.
func (s *Server) handleAuthProviders(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"password": true,
		"github":   s.githubAuth.Configured(),
	})
}

func (s *Server) redirectToApp(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, strings.TrimRight(s.cfg.AppURL, "/")+path, http.StatusFound)
}

func (s *Server) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
