package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	gh "github.com/NamanSharma2112/OpsPulse/services/api/internal/github"
)

// GitHubService runs the OAuth sign-in flow and hands out API clients that
// act as the signed-in user.
type GitHubService struct {
	oauth  gh.OAuthConfig
	users  UserStore
	tokens *TokenIssuer
	box    *SecretBox
}

// UserStore is the slice of user persistence this service needs.
type UserStore interface {
	UpsertGitHub(ctx context.Context, u *domain.User) error
}

// NewGitHubService wires the OAuth flow.
func NewGitHubService(oauth gh.OAuthConfig, users UserStore, tokens *TokenIssuer, box *SecretBox) *GitHubService {
	return &GitHubService{oauth: oauth, users: users, tokens: tokens, box: box}
}

// Configured reports whether sign-in with GitHub is available.
func (s *GitHubService) Configured() bool { return s.oauth.Configured() }

// AuthorizeURL is where the browser is sent to begin the flow.
func (s *GitHubService) AuthorizeURL(state string) string { return s.oauth.AuthorizeURL(state) }

// Complete exchanges the callback code for a session. The GitHub token is
// sealed before it reaches the database.
func (s *GitHubService) Complete(ctx context.Context, code string) (*Session, error) {
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	profile, err := gh.NewClient(token.AccessToken, s.oauth.APIBase).CurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("read github profile: %w", err)
	}
	if profile.ID == 0 {
		return nil, fmt.Errorf("%w: github returned no user id", domain.ErrInvalidInput)
	}

	sealed, err := s.box.Seal(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("seal github token: %w", err)
	}

	email := profile.Email
	if email == "" {
		// GitHub hides private addresses; synthesise the stable noreply form
		// rather than failing a sign-in over it.
		email = fmt.Sprintf("%d+%s@users.noreply.github.com", profile.ID, profile.Login)
	}
	name := profile.Name
	if name == "" {
		name = profile.Login
	}

	user := &domain.User{
		Email:             strings.ToLower(email),
		Name:              name,
		GitHubID:          &profile.ID,
		GitHubLogin:       profile.Login,
		AvatarURL:         profile.AvatarURL,
		GitHubAccessToken: sealed,
		GitHubTokenScopes: token.Scope,
	}
	if err := s.users.UpsertGitHub(ctx, user); err != nil {
		return nil, err
	}

	session, err := s.tokens.Issue(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &Session{
		Token:     session,
		ExpiresAt: nowUTC().Add(s.tokens.TTL()),
		User:      user,
	}, nil
}

// ClientFor returns a GitHub client acting as the given user, unsealing their
// stored token.
func (s *GitHubService) ClientFor(user *domain.User) (*gh.Client, error) {
	if !user.HasGitHubToken() {
		return nil, fmt.Errorf("%w: connect your GitHub account first", domain.ErrForbidden)
	}
	token, err := s.box.Open(user.GitHubAccessToken)
	if err != nil {
		return nil, fmt.Errorf("unseal github token: %w", err)
	}
	return gh.NewClient(token, s.oauth.APIBase), nil
}
