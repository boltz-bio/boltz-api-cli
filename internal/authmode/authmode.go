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

	refreshed, err := refreshSession(ctx, resolved)
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

func refreshSession(ctx context.Context, resolved authconfig.Resolved) (*authstore.Session, error) {
	session, err := authstore.WithLock(ctx, func(lockHeld bool) (*authstore.Session, error) {
		current, err := authstore.LoadSession()
		if err != nil {
			return nil, autherror.New("session_load_failed", "Failed to load OAuth session", err.Error())
		}
		if current == nil || !SessionMatches(resolved, *current) {
			return nil, sessionMismatchError(resolved)
		}
		if usableAccessToken(current) {
			return current, nil
		}

		providerMetadata := oauthclient.ProviderMetadata{
			IssuerURL:        current.IssuerURL,
			AuthorizationURL: coalesce(resolved.AuthorizationURL, current.AuthorizationURL),
			TokenURL:         coalesce(resolved.TokenURL, current.TokenURL),
			UserInfoURL:      coalesce(resolved.UserInfoURL, current.UserInfoURL),
			RevocationURL:    coalesce(resolved.RevocationURL, current.RevocationURL),
			JWKSURL:          strings.TrimSpace(current.JWKSURL),
			Algorithms:       append([]string(nil), current.Algorithms...),
		}
		if needsVerificationMetadata(providerMetadata) {
			discovered, discoveryErr := oauthclient.ResolveProviderMetadata(ctx, oauthclient.Config{
				IssuerURL:        current.IssuerURL,
				AuthorizationURL: providerMetadata.AuthorizationURL,
				TokenURL:         providerMetadata.TokenURL,
				UserInfoURL:      providerMetadata.UserInfoURL,
				RevocationURL:    providerMetadata.RevocationURL,
			})
			if discoveryErr == nil {
				providerMetadata.AuthorizationURL = coalesce(providerMetadata.AuthorizationURL, discovered.AuthorizationURL)
				providerMetadata.TokenURL = coalesce(providerMetadata.TokenURL, discovered.TokenURL)
				providerMetadata.UserInfoURL = coalesce(providerMetadata.UserInfoURL, discovered.UserInfoURL)
				providerMetadata.RevocationURL = coalesce(providerMetadata.RevocationURL, discovered.RevocationURL)
				providerMetadata.JWKSURL = coalesce(providerMetadata.JWKSURL, discovered.JWKSURL)
				if len(providerMetadata.Algorithms) == 0 {
					providerMetadata.Algorithms = append([]string(nil), discovered.Algorithms...)
				}
			}
		}

		refreshToken, backend, err := authstore.LoadRefreshTokenWithPreferredBackend(current.StorageBackend)
		if err != nil {
			return nil, autherror.New("refresh_token_load_failed", "Failed to load refresh token", err.Error())
		}
		if strings.TrimSpace(refreshToken) == "" {
			// Don't clear the session: an empty token can mean the keyring is
			// just unreachable from this process, and other processes may be
			// using the session.
			return nil, autherror.New("reauth_required", "Stored OAuth session can no longer be refreshed", "Could not load a refresh token (the system keyring may be unavailable). Run `boltz-api auth login` again.")
		}

		tokens, err := oauthclient.Refresh(ctx, oauthclient.RefreshConfig{
			Provider:     providerMetadata,
			ClientID:     current.ClientID,
			RefreshToken: refreshToken,
			Resource:     current.Audience,
		})
		if err != nil {
			if oauthclient.IsInvalidGrant(err) {
				// Without the lock, the rejection may mean a peer just spent
				// this token and is about to save its replacement.
				if !lockHeld {
					return nil, autherror.New("reauth_required", "Stored OAuth session is no longer valid", "Run `boltz-api auth login` again.")
				}
				if clearErr := authstore.ClearAll(); clearErr != nil {
					return nil, autherror.New("session_clear_failed", "Stored OAuth session is invalid and could not be cleared", clearErr.Error())
				}
				return nil, autherror.New("reauth_required", "Stored OAuth session is no longer valid", "Run `boltz-api auth login` again.")
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
		return nil, err
	}
	return session, nil
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
