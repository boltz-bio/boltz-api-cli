package authmode

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/autherror"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/boltz-bio/boltz-api-cli/internal/oauthclient"
)

const refreshThreshold = 60 * time.Second

var now = time.Now

type Mode string

const (
	ModeNone   Mode = "none"
	ModeAPIKey Mode = "api_key"
	ModeOAuth  Mode = "oauth"
)

type Result struct {
	Mode        Mode
	APIKey      string
	AccessToken string
	Session     *authstore.Session
}

func Resolve(ctx context.Context, resolved authconfig.Resolved) (Result, error) {
	if strings.TrimSpace(resolved.APIKey) != "" {
		return Result{Mode: ModeAPIKey, APIKey: strings.TrimSpace(resolved.APIKey)}, nil
	}

	session, err := authstore.LoadSession()
	if err != nil {
		return Result{}, autherror.New("session_load_failed", "Failed to load OAuth session", err.Error())
	}
	if session == nil {
		return Result{}, MissingLoginError()
	}
	if !SessionMatches(resolved, *session) {
		return Result{}, sessionMismatchError(resolved)
	}

	if usableAccessToken(session) {
		return Result{Mode: ModeOAuth, AccessToken: session.AccessToken, Session: session}, nil
	}

	refreshed, err := refreshSession(ctx, resolved, session)
	if err != nil {
		return Result{}, err
	}
	return Result{Mode: ModeOAuth, AccessToken: refreshed.AccessToken, Session: refreshed}, nil
}

func SessionMatches(resolved authconfig.Resolved, session authstore.Session) bool {
	if strings.TrimSpace(resolved.IssuerURL) == "" || strings.TrimSpace(resolved.ClientID) == "" {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(session.IssuerURL), strings.TrimSpace(resolved.IssuerURL)) {
		return false
	}
	if strings.TrimSpace(session.ClientID) != strings.TrimSpace(resolved.ClientID) {
		return false
	}
	if strings.TrimSpace(session.Audience) != strings.TrimSpace(resolved.Audience) {
		return false
	}
	return sameScopes(session.Scopes, resolved.Scopes)
}

// refreshSession refreshes an expiring access token.
//
// The session lock exists only to serialize refresh-token rotation between
// concurrent CLI processes: the server rotates refresh tokens, so two processes
// refreshing with the same token would make the loser's token invalid. Keep the
// lock as short as possible: OIDC discovery and the (potentially slow) keyring
// read happen before it is taken; under the lock we only re-read session.json,
// make the token request, and write the results.
//
// A failed refresh never deletes the stored session. Other processes may be
// mid-run on it, and the failure may be transient (keyring unreachable, lost
// rotation race, network). `auth login` overwrites it, `auth logout` clears it.
func refreshSession(ctx context.Context, resolved authconfig.Resolved, observed *authstore.Session) (*authstore.Session, error) {
	providerMetadata := resolveRefreshProvider(ctx, resolved, observed)
	refreshToken, backend, tokenErr := authstore.LoadRefreshTokenWithPreferredBackend(observed.StorageBackend)

	session, err := authstore.WithLock(ctx, func() (*authstore.Session, error) {
		current, err := authstore.LoadSession()
		if err != nil {
			return nil, autherror.New("session_load_failed", "Failed to load OAuth session", err.Error())
		}
		if current == nil {
			return nil, MissingLoginError()
		}
		if !SessionMatches(resolved, *current) {
			return nil, sessionMismatchError(resolved)
		}
		if usableAccessToken(current) {
			// Another process refreshed while we waited for the lock.
			return current, nil
		}
		if !sameSessionVersion(current, observed) {
			// session.json changed since we read the refresh token, so the token
			// we hold may already be rotated. Read it again under the lock.
			refreshToken, backend, tokenErr = authstore.LoadRefreshTokenWithPreferredBackend(current.StorageBackend)
		}
		if tokenErr != nil {
			return nil, refreshTokenLoadError(tokenErr)
		}
		if strings.TrimSpace(refreshToken) == "" {
			return nil, ReauthError("Stored OAuth session has no refresh token")
		}

		tokens, err := refreshWith(ctx, providerMetadata, current, refreshToken)
		if err != nil && oauthclient.IsInvalidGrant(err) {
			// The token may have been rotated by a process that wrote the new
			// refresh token but not yet the session (or crashed in between).
			// Re-read it and retry once if it changed.
			latest, latestBackend, latestErr := authstore.LoadRefreshTokenWithPreferredBackend(current.StorageBackend)
			if latestErr == nil && strings.TrimSpace(latest) != "" && latest != refreshToken {
				refreshToken, backend = latest, latestBackend
				tokens, err = refreshWith(ctx, providerMetadata, current, refreshToken)
			}
		}
		if err != nil {
			if oauthclient.IsInvalidGrant(err) {
				return nil, ReauthError("Stored OAuth session is no longer valid")
			}
			return nil, autherror.New("token_refresh_failed", "Failed to refresh OAuth access token", err.Error())
		}
		if strings.TrimSpace(tokens.AccessToken) == "" {
			return nil, autherror.New("token_refresh_failed", "Refresh response did not include an access token", "")
		}

		current.AccessToken = strings.TrimSpace(tokens.AccessToken)
		current.TokenType = coalesce(tokens.TokenType, current.TokenType)
		current.Expiry = tokens.Expiry
		current.IDToken = strings.TrimSpace(tokens.IDToken)
		if len(tokens.GrantedScopes) > 0 {
			current.GrantedScopes = append([]string(nil), tokens.GrantedScopes...)
		}
		if tokens.RefreshToken != "" {
			backend, err = authstore.SaveRefreshToken(tokens.RefreshToken)
			if err != nil {
				return nil, autherror.New("refresh_token_store_failed", "Failed to store rotated refresh token", err.Error())
			}
		}
		current.StorageBackend = backend
		current.AuthorizationURL = providerMetadata.AuthorizationURL
		current.JWKSURL = coalesce(current.JWKSURL, providerMetadata.JWKSURL)
		if len(current.Algorithms) == 0 && len(providerMetadata.Algorithms) > 0 {
			current.Algorithms = append([]string(nil), providerMetadata.Algorithms...)
		}
		current.TokenURL = coalesce(resolved.TokenURL, current.TokenURL)
		current.UserInfoURL = coalesce(resolved.UserInfoURL, current.UserInfoURL)
		current.RevocationURL = coalesce(resolved.RevocationURL, current.RevocationURL)

		if err := authstore.SaveSession(*current); err != nil {
			return nil, autherror.New("session_save_failed", "Failed to persist refreshed OAuth session", err.Error())
		}
		return current, nil
	})
	if err != nil {
		var lockErr *authstore.LockTimeoutError
		if errors.As(err, &lockErr) {
			// The holder may have finished a refresh just before we gave up.
			if latest, loadErr := authstore.LoadSession(); loadErr == nil && latest != nil && SessionMatches(resolved, *latest) && usableAccessToken(latest) {
				return latest, nil
			}
			return nil, lockTimeoutError(lockErr)
		}
		return nil, err
	}
	return session, nil
}

func refreshWith(ctx context.Context, provider oauthclient.ProviderMetadata, session *authstore.Session, refreshToken string) (*oauthclient.TokenSet, error) {
	return oauthclient.Refresh(ctx, oauthclient.RefreshConfig{
		Provider:     provider,
		ClientID:     session.ClientID,
		RefreshToken: refreshToken,
		Resource:     session.Audience,
	})
}

// resolveRefreshProvider builds the provider metadata for a refresh, running
// OIDC discovery for legacy sessions that lack verification metadata. It runs
// outside the session lock because it is network I/O that does not depend on
// the refresh token.
func resolveRefreshProvider(ctx context.Context, resolved authconfig.Resolved, session *authstore.Session) oauthclient.ProviderMetadata {
	providerMetadata := oauthclient.ProviderMetadata{
		IssuerURL:        session.IssuerURL,
		AuthorizationURL: coalesce(resolved.AuthorizationURL, session.AuthorizationURL),
		TokenURL:         coalesce(resolved.TokenURL, session.TokenURL),
		UserInfoURL:      coalesce(resolved.UserInfoURL, session.UserInfoURL),
		RevocationURL:    coalesce(resolved.RevocationURL, session.RevocationURL),
		JWKSURL:          strings.TrimSpace(session.JWKSURL),
		Algorithms:       append([]string(nil), session.Algorithms...),
	}
	if !needsVerificationMetadata(providerMetadata) {
		return providerMetadata
	}
	discovered, err := oauthclient.ResolveProviderMetadata(ctx, oauthclient.Config{
		IssuerURL:        session.IssuerURL,
		AuthorizationURL: providerMetadata.AuthorizationURL,
		TokenURL:         providerMetadata.TokenURL,
		UserInfoURL:      providerMetadata.UserInfoURL,
		RevocationURL:    providerMetadata.RevocationURL,
	})
	if err != nil {
		return providerMetadata
	}
	providerMetadata.AuthorizationURL = coalesce(providerMetadata.AuthorizationURL, discovered.AuthorizationURL)
	providerMetadata.TokenURL = coalesce(providerMetadata.TokenURL, discovered.TokenURL)
	providerMetadata.UserInfoURL = coalesce(providerMetadata.UserInfoURL, discovered.UserInfoURL)
	providerMetadata.RevocationURL = coalesce(providerMetadata.RevocationURL, discovered.RevocationURL)
	providerMetadata.JWKSURL = coalesce(providerMetadata.JWKSURL, discovered.JWKSURL)
	if len(providerMetadata.Algorithms) == 0 {
		providerMetadata.Algorithms = append([]string(nil), discovered.Algorithms...)
	}
	return providerMetadata
}

// sameSessionVersion reports whether session.json still holds what we read
// before taking the lock. Every token rotation rewrites the session, so an
// unchanged session means the refresh token read earlier is still current.
func sameSessionVersion(a, b *authstore.Session) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.AccessToken == b.AccessToken &&
		a.UpdatedAt.Equal(b.UpdatedAt) &&
		a.StorageBackend == b.StorageBackend
}

func refreshTokenLoadError(err error) error {
	if errors.Is(err, authstore.ErrKeyringUnavailable) || errors.Is(err, authstore.ErrKeyringTimeout) {
		return autherror.NewWithHints(
			"keyring_unavailable",
			"Could not read the refresh token from the system keyring",
			err.Error()+". The stored session was kept; retry once the keyring is reachable.",
			"On headless hosts, set `"+authstore.DisableKeyringEnv+"=1` and run `boltz-api auth login` to store the refresh token in a 0600 file instead.",
			"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode instead.",
		)
	}
	return autherror.New("refresh_token_load_failed", "Failed to load refresh token", err.Error())
}

func lockTimeoutError(err *authstore.LockTimeoutError) error {
	return autherror.NewWithHints(
		"session_lock_timeout",
		"Timed out waiting for another boltz-api process to finish refreshing the OAuth session",
		err.Error()+".",
		"If that process is stuck (often on a keyring prompt), stop it and retry.",
		"Raise the wait with `"+authstore.LockTimeoutEnv+"` (for example `120s`).",
		"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode, which never takes the lock.",
	)
}

func needsVerificationMetadata(provider oauthclient.ProviderMetadata) bool {
	return strings.TrimSpace(provider.JWKSURL) == "" || len(provider.Algorithms) == 0
}

func usableAccessToken(session *authstore.Session) bool {
	if session == nil {
		return false
	}
	if strings.TrimSpace(session.AccessToken) == "" {
		return false
	}
	if session.Expiry.IsZero() {
		return true
	}
	return session.Expiry.After(now().UTC().Add(refreshThreshold))
}

func sameScopes(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	a := append([]string(nil), left...)
	b := append([]string(nil), right...)
	for i := range a {
		a[i] = strings.TrimSpace(a[i])
	}
	for i := range b {
		b[i] = strings.TrimSpace(b[i])
	}
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func IsAuthError(err error) bool {
	var authErr *autherror.Error
	return errors.As(err, &authErr)
}

func WrapConfigError(err error) error {
	if err == nil {
		return nil
	}
	return autherror.NewWithHints(
		"config_load_failed",
		"Failed to load auth configuration",
		err.Error(),
		"Fix the local auth config file and try again.",
	)
}

func MissingLoginError() error {
	return autherror.NewWithHints(
		"auth_required",
		"Authentication required",
		"Run `boltz-api auth login` to create an OAuth session.",
		"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode.",
		"Run `boltz-api auth whoami` to inspect the current local auth state.",
	)
}

func ReauthError(message string) error {
	if strings.TrimSpace(message) == "" {
		message = "Stored OAuth session is no longer valid"
	}
	return autherror.NewWithHints(
		"reauth_required",
		message,
		"Run `boltz-api auth login` again to refresh the local session.",
		"Run `boltz-api auth whoami` to inspect the local auth state.",
		"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode instead.",
	)
}

func DescribeMode(resolved authconfig.Resolved, session *authstore.Session) Mode {
	if strings.TrimSpace(resolved.APIKey) != "" {
		return ModeAPIKey
	}
	if session != nil && SessionMatches(resolved, *session) {
		return ModeOAuth
	}
	return ModeNone
}

func MatchError(resolved authconfig.Resolved, session *authstore.Session) error {
	if session == nil {
		return MissingLoginError()
	}
	if !SessionMatches(resolved, *session) {
		return sessionMismatchError(resolved)
	}
	return nil
}

func sessionMismatchError(resolved authconfig.Resolved) error {
	return autherror.NewWithHints(
		"session_mismatch",
		"Stored OAuth session does not match the current auth settings",
		fmt.Sprintf("Expected issuer %q and client %q.", resolved.IssuerURL, resolved.ClientID),
		"Run `boltz-api auth login` again to create a matching OAuth session.",
		"Run `boltz-api auth whoami` to inspect the local auth state.",
		"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode instead.",
	)
}
