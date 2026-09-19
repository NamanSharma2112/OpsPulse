package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

// Service handles registration and sign-in.
type Service struct {
	users  store.Users
	tokens *TokenIssuer
}

// NewService wires the auth service.
func NewService(users store.Users, tokens *TokenIssuer) *Service {
	return &Service{users: users, tokens: tokens}
}

// Session is what a successful sign-in returns to the dashboard.
type Session struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *domain.User `json:"user"`
}

// Register creates a password account and signs the caller straight in.
func (s *Service) Register(ctx context.Context, email, name, password string) (*Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: a valid email is required", domain.ErrInvalidInput)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidInput)
	}
	if name == "" {
		name = strings.SplitN(email, "@", 2)[0]
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{Email: email, Name: name, PasswordHash: hash}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return s.newSession(user)
}

// Login verifies a password and returns a session.
func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	user, err := s.users.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Do not leak which half of the pair was wrong.
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	if user.PasswordHash == "" {
		return nil, fmt.Errorf("%w: this account signs in with GitHub", domain.ErrUnauthorized)
	}
	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrUnauthorized
	}
	return s.newSession(user)
}

// LoginWithGitHub links or refreshes an account from a GitHub identity.
func (s *Service) LoginWithGitHub(ctx context.Context, githubID int64, login, name, email, avatarURL string) (*Session, error) {
	if githubID == 0 {
		return nil, fmt.Errorf("%w: missing GitHub user id", domain.ErrInvalidInput)
	}
	if email == "" {
		// GitHub hides private addresses; synthesise a stable placeholder.
		email = fmt.Sprintf("%s@users.noreply.github.com", login)
	}
	if name == "" {
		name = login
	}
	user := &domain.User{
		Email:       strings.ToLower(email),
		Name:        name,
		GitHubID:    &githubID,
		GitHubLogin: login,
		AvatarURL:   avatarURL,
	}
	if err := s.users.UpsertGitHub(ctx, user); err != nil {
		return nil, err
	}
	return s.newSession(user)
}

// Authenticate resolves the user behind a bearer token.
func (s *Service) Authenticate(ctx context.Context, token string) (*domain.User, error) {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	return user, nil
}

// nowUTC exists so the GitHub flow and the password flow agree on the clock.
func nowUTC() time.Time { return time.Now().UTC() }

func (s *Service) newSession(user *domain.User) (*Session, error) {
	token, err := s.tokens.Issue(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &Session{
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(s.tokens.TTL()),
		User:      user,
	}, nil
}
