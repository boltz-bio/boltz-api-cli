// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/autherror"
	"github.com/boltz-bio/boltz-api-cli/internal/authmode"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/boltz-bio/boltz-api-cli/internal/oauthclient"
	computeapi "github.com/boltz-bio/boltz-api-go"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

const authStatusWarningScopeMismatch = "Granted OAuth scopes do not include the full configured scope set."

const (
	defaultAuthWaitTimeout      = time.Minute
	defaultAuthWaitPollInterval = 2 * time.Second
)

var authNow = time.Now

var authCommand = &cli.Command{
	Name:            "auth",
	Usage:           "Manage CLI authentication",
	Suggest:         true,
	HideHelpCommand: true,
	Commands: []*cli.Command{
		{
			Name:            "login",
			Usage:           "Run the OAuth login flow and persist the resulting session",
			Action:          handleAuthLogin,
			HideHelpCommand: true,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "device-code",
					Usage: "Use OAuth device authorization instead of a localhost browser callback",
				},
				&cli.BoolFlag{
					Name:  "json-events",
					Usage: "Emit newline-delimited JSON progress events for device-code login",
				},
			},
		},
		{
			Name:            "logout",
			Usage:           "Clear the local OAuth session and refresh token",
			Action:          handleAuthLogout,
			HideHelpCommand: true,
		},
		{
			Name:            "whoami",
			Usage:           "Show the effective local authentication state",
			Action:          handleAuthWhoAmI,
			HideHelpCommand: true,
		},
		{
			Name:            "status",
			Usage:           "Show stable machine-readable auth status without refreshing tokens",
			Action:          handleAuthStatus,
			HideHelpCommand: true,
		},
		{
			Name:            "validate",
			Usage:           "Validate local auth state and refresh OAuth sessions when needed",
			Action:          handleAuthValidate,
			HideHelpCommand: true,
		},
		{
			Name:            "orgs",
			Aliases:         []string{"organizations"},
			Usage:           "List organizations available to the current credential",
			Action:          handleAuthOrgs,
			HideHelpCommand: true,
		},
		{
			Name:            "wait",
			Usage:           "Wait for usable local auth to become available",
			Action:          handleAuthWait,
			HideHelpCommand: true,
			Flags: []cli.Flag{
				&cli.DurationFlag{
					Name:  "timeout",
					Value: defaultAuthWaitTimeout,
					Usage: "How long to wait before returning a waiting status",
				},
				&cli.DurationFlag{
					Name:  "poll-interval",
					Value: defaultAuthWaitPollInterval,
					Usage: "How often to re-check local auth state while waiting",
				},
			},
		},
		{
			Name:            "switch-org",
			Usage:           "Persist a local selected organization for OAuth requests",
			ArgsUsage:       "<org>",
			Action:          handleAuthSwitchOrg,
			HideHelpCommand: true,
		},
	},
}

type whoAmIResponse struct {
	Mode                string          `json:"mode"`
	APIKeyConfigured    bool            `json:"api_key_configured"`
	IssuerURL           string          `json:"issuer_url,omitempty"`
	ClientID            string          `json:"client_id,omitempty"`
	Audience            string          `json:"audience,omitempty"`
	Scopes              []string        `json:"scopes,omitempty"`
	SelectedOrg         string          `json:"selected_org,omitempty"`
	StorageBackend      string          `json:"storage_backend,omitempty"`
	SessionMatch        bool            `json:"session_match"`
	RefreshTokenPresent bool            `json:"refresh_token_present"`
	Expiry              *time.Time      `json:"expiry,omitempty"`
	Identity            *whoAmIIdentity `json:"identity,omitempty"`
	Sources             whoAmISources   `json:"sources,omitempty"`
}

type whoAmIIdentity struct {
	Subject           string `json:"subject,omitempty"`
	Email             string `json:"email,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

type whoAmISources struct {
	APIKey       authconfig.Source `json:"api_key,omitempty"`
	IssuerURL    authconfig.Source `json:"issuer_url,omitempty"`
	ClientID     authconfig.Source `json:"client_id,omitempty"`
	Scopes       authconfig.Source `json:"scopes,omitempty"`
	Audience     authconfig.Source `json:"audience,omitempty"`
	SelectedOrg  authconfig.Source `json:"selected_org,omitempty"`
	Session      authconfig.Source `json:"session,omitempty"`
	RefreshToken authconfig.Source `json:"refresh_token,omitempty"`
}

type authStatusResponse struct {
	Authenticated        bool                 `json:"authenticated"`
	EffectiveMode        string               `json:"effective_mode"`
	ActiveSource         string               `json:"active_source"`
	Config               authConfigDetails    `json:"config"`
	APIKeyConfigured     bool                 `json:"api_key_configured"`
	APIKeyOverridesOAuth bool                 `json:"api_key_overrides_oauth"`
	IssuerURL            string               `json:"issuer_url,omitempty"`
	ClientID             string               `json:"client_id,omitempty"`
	Audience             string               `json:"audience,omitempty"`
	SelectedOrg          string               `json:"selected_org,omitempty"`
	RequestedScopes      []string             `json:"requested_scopes,omitempty"`
	GrantedScopes        []string             `json:"granted_scopes,omitempty"`
	GrantedScopesKnown   bool                 `json:"granted_scopes_known"`
	MissingScopes        []string             `json:"missing_scopes,omitempty"`
	SessionPresent       bool                 `json:"session_present"`
	SessionMatch         bool                 `json:"session_match"`
	RefreshTokenPresent  bool                 `json:"refresh_token_present"`
	Refreshable          bool                 `json:"refreshable"`
	StorageBackend       string               `json:"storage_backend,omitempty"`
	Expiry               *time.Time           `json:"expiry,omitempty"`
	Identity             *whoAmIIdentity      `json:"identity,omitempty"`
	StoredOAuthSession   *oauthSessionDetails `json:"stored_oauth_session,omitempty"`
	Sources              whoAmISources        `json:"sources,omitempty"`
	Warnings             []string             `json:"warnings,omitempty"`
	Actions              []string             `json:"actions,omitempty"`
}

type authValidateResponse struct {
	Valid               bool                 `json:"valid"`
	EffectiveMode       string               `json:"effective_mode"`
	ActiveSource        string               `json:"active_source"`
	Refreshed           bool                 `json:"refreshed"`
	Config              authConfigDetails    `json:"config"`
	IssuerURL           string               `json:"issuer_url,omitempty"`
	ClientID            string               `json:"client_id,omitempty"`
	Audience            string               `json:"audience,omitempty"`
	SelectedOrg         string               `json:"selected_org,omitempty"`
	RequestedScopes     []string             `json:"requested_scopes,omitempty"`
	GrantedScopes       []string             `json:"granted_scopes,omitempty"`
	GrantedScopesKnown  bool                 `json:"granted_scopes_known"`
	MissingScopes       []string             `json:"missing_scopes,omitempty"`
	Refreshable         bool                 `json:"refreshable"`
	RefreshTokenPresent bool                 `json:"refresh_token_present"`
	SessionMatch        bool                 `json:"session_match"`
	StorageBackend      string               `json:"storage_backend,omitempty"`
	Expiry              *time.Time           `json:"expiry,omitempty"`
	Identity            *whoAmIIdentity      `json:"identity,omitempty"`
	StoredOAuthSession  *oauthSessionDetails `json:"stored_oauth_session,omitempty"`
	Warnings            []string             `json:"warnings,omitempty"`
	Actions             []string             `json:"actions,omitempty"`
}

type authMeResponse struct {
	PrincipalType           string                     `json:"principal_type"`
	UserID                  string                     `json:"user_id,omitempty"`
	APIKeyID                string                     `json:"api_key_id,omitempty"`
	KeyType                 string                     `json:"key_type,omitempty"`
	Mode                    string                     `json:"mode,omitempty"`
	OrganizationID          string                     `json:"organization_id,omitempty"`
	SelectedOrganizationID  string                     `json:"selected_organization_id,omitempty"`
	ActiveOrganizationID    string                     `json:"active_organization_id,omitempty"`
	WorkspaceID             *string                    `json:"workspace_id,omitempty"`
	OrganizationMemberships []authMeOrganizationMember `json:"organization_memberships,omitempty"`
}

type authConfigDetails struct {
	Path    string `json:"path,omitempty"`
	Present bool   `json:"present"`
}

type authMeOrganizationMember struct {
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

type authOrganizationsResponse struct {
	PrincipalType string                    `json:"principal_type"`
	SelectedOrg   string                    `json:"selected_org,omitempty"`
	Organizations []authOrganizationSummary `json:"organizations"`
	APIKey        *authAPIKeySummary        `json:"api_key,omitempty"`
}

type authOrganizationSummary struct {
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role,omitempty"`
	Selected       bool   `json:"selected"`
	Switchable     bool   `json:"switchable"`
}

type authAPIKeySummary struct {
	APIKeyID    string  `json:"api_key_id"`
	KeyType     string  `json:"key_type"`
	Mode        string  `json:"mode"`
	WorkspaceID *string `json:"workspace_id,omitempty"`
}

type authWaitResponse struct {
	authStatusResponse
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
	TimedOut     bool   `json:"timed_out,omitempty"`
	Timeout      string `json:"timeout,omitempty"`
	PollInterval string `json:"poll_interval,omitempty"`
}

type oauthSessionDetails struct {
	GrantedScopes       []string        `json:"granted_scopes,omitempty"`
	GrantedScopesKnown  bool            `json:"granted_scopes_known"`
	MissingScopes       []string        `json:"missing_scopes,omitempty"`
	RefreshTokenPresent bool            `json:"refresh_token_present"`
	Refreshable         bool            `json:"refreshable"`
	StorageBackend      string          `json:"storage_backend,omitempty"`
	Expiry              *time.Time      `json:"expiry,omitempty"`
	Identity            *whoAmIIdentity `json:"identity,omitempty"`
}

func handleAuthLogin(ctx context.Context, cmd *cli.Command) error {
	resolved, err := authconfig.Resolve(cmd)
	if err != nil {
		return authmode.WrapConfigError(err)
	}
	if strings.TrimSpace(resolved.IssuerURL) == "" {
		return autherror.New("missing_issuer_url", "OAuth issuer URL is required", "Set `--auth-issuer-url` or `BOLTZ_API_AUTH_ISSUER_URL`.")
	}
	if strings.TrimSpace(resolved.ClientID) == "" {
		return autherror.New("missing_client_id", "OAuth client ID is required", "Set `--auth-client-id` or `BOLTZ_API_AUTH_CLIENT_ID`.")
	}
	if cmd.Bool("json-events") && !cmd.Bool("device-code") {
		return autherror.New("unsupported_login_output", "`--json-events` requires `--device-code`", "Use `boltz-api auth login --device-code --json-events`.")
	}

	writer := commandWriter(cmd)
	jsonEvents := newAuthLoginEventWriter(writer, cmd.Bool("json-events"))

	loginConfig := oauthclient.Config{
		IssuerURL:        resolved.IssuerURL,
		ClientID:         resolved.ClientID,
		Scopes:           resolved.Scopes,
		Audience:         resolved.Audience,
		AuthorizationURL: resolved.AuthorizationURL,
		TokenURL:         resolved.TokenURL,
		UserInfoURL:      resolved.UserInfoURL,
		RevocationURL:    resolved.RevocationURL,
		ListenPort:       resolved.ListenPort,
		Output:           writer,
		OnDeviceCode:     jsonEvents.deviceCode,
	}
	if jsonEvents.enabled {
		loginConfig.Output = nil
	}

	var result *oauthclient.LoginResult
	if cmd.Bool("device-code") {
		result, err = oauthclient.DeviceLogin(ctx, loginConfig)
	} else {
		result, err = oauthclient.Login(ctx, loginConfig)
	}
	if err != nil {
		return autherror.New("login_failed", "OAuth login failed", err.Error())
	}

	_, err = authstore.WithLock(ctx, func() (struct{}, error) {
		previousSession, loadErr := authstore.LoadSession()
		if loadErr != nil {
			return struct{}{}, autherror.New("session_load_failed", "Failed to load the existing OAuth session", loadErr.Error())
		}
		previousRefreshToken, _, loadErr := authstore.LoadRefreshToken()
		if loadErr != nil {
			return struct{}{}, autherror.New("refresh_token_load_failed", "Failed to load the existing refresh token", loadErr.Error())
		}

		rollback := func(code, message string, cause error) error {
			restoreErr := restoreLoginState(previousSession, previousRefreshToken)
			if restoreErr != nil {
				return autherror.New(code, message, cause.Error()+". Rollback failed: "+restoreErr.Error())
			}
			return autherror.New(code, message, cause.Error())
		}

		backend := ""
		if strings.TrimSpace(result.Tokens.RefreshToken) != "" {
			var storeErr error
			backend, storeErr = authstore.SaveRefreshToken(result.Tokens.RefreshToken)
			if storeErr != nil {
				return struct{}{}, rollback("refresh_token_store_failed", "Failed to store refresh token", storeErr)
			}
		} else {
			if clearErr := authstore.ClearRefreshToken(); clearErr != nil {
				return struct{}{}, rollback("refresh_token_store_failed", "Failed to clear stale refresh token", clearErr)
			}
		}

		grantedScopes := append([]string(nil), result.Tokens.GrantedScopes...)
		if len(grantedScopes) == 0 {
			grantedScopes = append([]string(nil), resolved.Scopes...)
		}

		session := authstore.Session{
			IssuerURL:        result.Provider.IssuerURL,
			ClientID:         resolved.ClientID,
			Audience:         resolved.Audience,
			Scopes:           append([]string(nil), resolved.Scopes...),
			GrantedScopes:    grantedScopes,
			AccessToken:      result.Tokens.AccessToken,
			TokenType:        result.Tokens.TokenType,
			Expiry:           result.Tokens.Expiry,
			IDToken:          result.Tokens.IDToken,
			AuthorizationURL: result.Provider.AuthorizationURL,
			TokenURL:         result.Provider.TokenURL,
			UserInfoURL:      result.Provider.UserInfoURL,
			RevocationURL:    result.Provider.RevocationURL,
			JWKSURL:          result.Provider.JWKSURL,
			Algorithms:       append([]string(nil), result.Provider.Algorithms...),
			StorageBackend:   backend,
			Identity:         result.Identity,
		}
		if err := authstore.SaveSession(session); err != nil {
			return struct{}{}, rollback("session_save_failed", "Failed to persist OAuth session", err)
		}
		if err := authconfig.SaveProfile(resolved); err != nil {
			return struct{}{}, rollback("config_save_failed", "Failed to persist auth configuration", err)
		}
		return struct{}{}, nil
	})
	if err != nil {
		return err
	}

	if jsonEvents.enabled {
		if err := jsonEvents.write(map[string]any{"event": "success"}); err != nil {
			return err
		}
		if strings.TrimSpace(resolved.APIKey) != "" {
			return jsonEvents.write(map[string]any{
				"event":   "warning",
				"message": "API-key mode is still active for commands in this shell. Clear `--api-key` or `BOLTZ_API_KEY` to use the stored OAuth session.",
			})
		}
		return nil
	}

	fmt.Fprintln(writer, "Authentication successful.")
	if strings.TrimSpace(resolved.APIKey) != "" {
		fmt.Fprintln(writer, "API-key mode is still active for commands in this shell. Clear `--api-key` or `BOLTZ_API_KEY` to use the stored OAuth session.")
	}
	return nil
}

type authLoginEventWriter struct {
	encoder *json.Encoder
	enabled bool
}

func newAuthLoginEventWriter(w io.Writer, enabled bool) authLoginEventWriter {
	if !enabled {
		return authLoginEventWriter{}
	}
	return authLoginEventWriter{
		encoder: json.NewEncoder(w),
		enabled: true,
	}
}

func (w authLoginEventWriter) deviceCode(device oauthclient.DeviceAuthorization) {
	if !w.enabled {
		return
	}

	url := strings.TrimSpace(device.VerificationURIComplete)
	if url == "" {
		url = strings.TrimSpace(device.VerificationURI)
	}
	_ = w.write(map[string]any{
		"event":                     "auth_url",
		"url":                       url,
		"verification_uri":          strings.TrimSpace(device.VerificationURI),
		"verification_uri_complete": strings.TrimSpace(device.VerificationURIComplete),
		"user_code":                 strings.TrimSpace(device.UserCode),
		"expires_in":                device.ExpiresIn,
		"interval":                  device.Interval,
	})
}

func (w authLoginEventWriter) write(event map[string]any) error {
	if !w.enabled {
		return nil
	}
	return w.encoder.Encode(event)
}

func handleAuthLogout(ctx context.Context, cmd *cli.Command) error {
	resolved, err := authconfig.Resolve(cmd)
	if err != nil {
		return authmode.WrapConfigError(err)
	}

	type logoutSnapshot struct {
		session      *authstore.Session
		refreshToken string
	}

	snapshot, err := authstore.WithLock(ctx, func() (logoutSnapshot, error) {
		session, err := authstore.LoadSession()
		if err != nil {
			return logoutSnapshot{}, autherror.New("session_load_failed", "Failed to load OAuth session", err.Error())
		}
		refreshToken, _, err := authstore.LoadRefreshToken()
		if err != nil {
			return logoutSnapshot{}, autherror.New("refresh_token_load_failed", "Failed to load refresh token", err.Error())
		}
		snapshot := logoutSnapshot{
			session:      session,
			refreshToken: refreshToken,
		}
		if err := authstore.ClearAll(); err != nil {
			return logoutSnapshot{}, autherror.New("logout_failed", "Failed to clear the local OAuth session", err.Error())
		}
		return snapshot, nil
	})
	if err != nil {
		return err
	}

	if snapshot.session != nil {
		endpoint := snapshot.session.RevocationURL
		if strings.TrimSpace(resolved.RevocationURL) != "" {
			endpoint = resolved.RevocationURL
		}
		tokenToRevoke := strings.TrimSpace(snapshot.refreshToken)
		tokenHint := "refresh_token"
		if tokenToRevoke == "" {
			tokenToRevoke = strings.TrimSpace(snapshot.session.AccessToken)
			tokenHint = "access_token"
		}
		_ = oauthclient.Revoke(ctx, oauthclient.RevokeConfig{
			Endpoint:      endpoint,
			ClientID:      snapshot.session.ClientID,
			Token:         tokenToRevoke,
			TokenTypeHint: tokenHint,
		})
	}

	if strings.TrimSpace(resolved.APIKey) != "" {
		fmt.Fprintln(commandWriter(cmd), "Logged out. Local OAuth state was cleared, but API-key mode is still active via `--api-key` or `BOLTZ_API_KEY`.")
		return nil
	}

	fmt.Fprintln(commandWriter(cmd), "Logged out.")
	return nil
}

func handleAuthWhoAmI(ctx context.Context, cmd *cli.Command) error {
	_ = ctx

	snapshot, err := loadAuthSnapshot(cmd)
	if err != nil {
		return err
	}

	response := whoAmIResponse{
		Mode:                string(snapshot.mode),
		APIKeyConfigured:    strings.TrimSpace(snapshot.resolved.APIKey) != "",
		IssuerURL:           snapshot.resolved.IssuerURL,
		ClientID:            snapshot.resolved.ClientID,
		Audience:            snapshot.resolved.Audience,
		Scopes:              append([]string(nil), snapshot.resolved.Scopes...),
		SelectedOrg:         snapshot.resolved.SelectedOrg,
		SessionMatch:        snapshot.sessionMatch,
		RefreshTokenPresent: strings.TrimSpace(snapshot.refreshToken) != "",
		Sources:             snapshot.sources,
	}

	if snapshot.session != nil {
		if !snapshot.session.Expiry.IsZero() {
			expiry := snapshot.session.Expiry
			response.Expiry = &expiry
		}
		response.Identity = summarizeIdentity(snapshot.session.Identity)
		response.StorageBackend = snapshot.session.StorageBackend
	}
	if response.StorageBackend == "" {
		response.StorageBackend = snapshot.refreshBackend
	}

	return showJSONValue(cmd, response, "auth whoami")
}

func handleAuthStatus(ctx context.Context, cmd *cli.Command) error {
	_ = ctx

	snapshot, err := loadAuthSnapshot(cmd)
	if err != nil {
		return err
	}

	response, ok := buildAuthStatusResponse(snapshot)
	if err := showJSONValue(cmd, response, "auth status"); err != nil {
		return err
	}
	if !ok {
		return cli.Exit("", 1)
	}
	return nil
}

func handleAuthValidate(ctx context.Context, cmd *cli.Command) error {
	snapshotBefore, err := loadAuthSnapshot(cmd)
	if err != nil {
		return err
	}

	result, err := authmode.Resolve(ctx, snapshotBefore.resolved)
	if err != nil {
		snapshotAfter, loadErr := loadAuthSnapshot(cmd)
		if loadErr != nil {
			return loadErr
		}

		response, _ := buildAuthStatusResponse(snapshotAfter)
		validate := authValidateResponse{
			Valid:               false,
			EffectiveMode:       response.EffectiveMode,
			ActiveSource:        response.ActiveSource,
			Config:              response.Config,
			IssuerURL:           response.IssuerURL,
			ClientID:            response.ClientID,
			Audience:            response.Audience,
			SelectedOrg:         response.SelectedOrg,
			RequestedScopes:     response.RequestedScopes,
			GrantedScopes:       response.GrantedScopes,
			GrantedScopesKnown:  response.GrantedScopesKnown,
			MissingScopes:       response.MissingScopes,
			Refreshable:         response.Refreshable,
			RefreshTokenPresent: response.RefreshTokenPresent,
			SessionMatch:        response.SessionMatch,
			StorageBackend:      response.StorageBackend,
			Expiry:              response.Expiry,
			Identity:            response.Identity,
			StoredOAuthSession:  response.StoredOAuthSession,
			Warnings:            response.Warnings,
			Actions:             append([]string(nil), response.Actions...),
		}
		if authErr, ok := err.(*autherror.Error); ok {
			validate.Actions = mergeStrings(validate.Actions, authErr.Envelope().Hints)
		} else if err.Error() != "" {
			validate.Actions = mergeStrings(validate.Actions, []string{err.Error()})
		}

		if showErr := showJSONValue(cmd, validate, "auth validate"); showErr != nil {
			return showErr
		}
		return cli.Exit("", 1)
	}

	snapshotAfter, err := loadAuthSnapshot(cmd)
	if err != nil {
		return err
	}
	response, _ := buildAuthStatusResponse(snapshotAfter)
	validate := authValidateResponse{
		Valid:               true,
		EffectiveMode:       string(result.Mode),
		ActiveSource:        response.ActiveSource,
		Refreshed:           wasSessionRefreshed(snapshotBefore.session, snapshotAfter.session),
		Config:              response.Config,
		IssuerURL:           response.IssuerURL,
		ClientID:            response.ClientID,
		Audience:            response.Audience,
		SelectedOrg:         response.SelectedOrg,
		RequestedScopes:     response.RequestedScopes,
		GrantedScopes:       response.GrantedScopes,
		GrantedScopesKnown:  response.GrantedScopesKnown,
		MissingScopes:       response.MissingScopes,
		Refreshable:         response.Refreshable,
		RefreshTokenPresent: response.RefreshTokenPresent,
		SessionMatch:        response.SessionMatch,
		StorageBackend:      response.StorageBackend,
		Expiry:              response.Expiry,
		Identity:            response.Identity,
		StoredOAuthSession:  response.StoredOAuthSession,
		Warnings:            response.Warnings,
		Actions:             response.Actions,
	}
	applyValidateModeNotes(&validate, result.Mode)
	return showJSONValue(cmd, validate, "auth validate")
}

func handleAuthOrgs(ctx context.Context, cmd *cli.Command) error {
	me, err := fetchAuthMe(ctx, cmd)
	if err != nil {
		return autherror.New("auth_context_failed", "Failed to load compute auth context", err.Error())
	}

	return showJSONValue(cmd, buildAuthOrganizationsResponse(me), "auth orgs")
}

func handleAuthWait(ctx context.Context, cmd *cli.Command) error {
	timeout := cmd.Duration("timeout")
	pollInterval := cmd.Duration("poll-interval")
	if timeout < 0 {
		return autherror.New("invalid_timeout", "Timeout must be zero or greater", "Set `--timeout` to a positive duration like `60s`.")
	}
	if pollInterval <= 0 {
		return autherror.New(
			"invalid_poll_interval",
			"Poll interval must be greater than zero",
			"Set `--poll-interval` to a positive duration like `2s`.",
		)
	}

	deadline := authNow().UTC().Add(timeout)

	for {
		snapshot, err := loadAuthSnapshot(cmd)
		if err != nil {
			return err
		}

		status, authenticated := buildAuthStatusResponse(snapshot)
		response := authWaitResponse{
			authStatusResponse: status,
			Status:             "waiting",
			Message:            "Usable local authentication is not ready yet.",
			Timeout:            timeout.String(),
			PollInterval:       pollInterval.String(),
		}
		if authenticated {
			response.Status = "success"
			response.Message = "Usable local authentication is available."
			return showJSONValue(cmd, response, "auth wait")
		}

		remaining := deadline.Sub(authNow().UTC())
		if timeout == 0 || remaining <= 0 {
			response.TimedOut = true
			response.Message = "Timed out waiting for usable local authentication."
			if err := showJSONValue(cmd, response, "auth wait"); err != nil {
				return err
			}
			return cli.Exit("", 1)
		}

		sleepFor := pollInterval
		if remaining < sleepFor {
			sleepFor = remaining
		}
		timer := time.NewTimer(sleepFor)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func handleAuthSwitchOrg(ctx context.Context, cmd *cli.Command) error {
	_ = ctx

	args := cmd.Args().Slice()
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return autherror.New("missing_org", "An organization value is required", "Usage: `boltz-api auth switch-org <org>`.")
	}

	org := strings.TrimSpace(args[0])
	if err := authconfig.SaveSelectedOrg(org); err != nil {
		return autherror.New("config_save_failed", "Failed to persist the selected organization", err.Error())
	}

	return showJSONValue(cmd, map[string]string{"selected_org": org}, "auth switch-org")
}

func fetchAuthMe(ctx context.Context, cmd *cli.Command) (authMeResponse, error) {
	client := computeapi.NewClient(getDefaultRequestOptions(cmd)...)
	var response authMeResponse
	if err := client.Get(ctx, "/compute/v1/auth/me", nil, &response); err != nil {
		return authMeResponse{}, err
	}
	return response, nil
}

func buildAuthOrganizationsResponse(me authMeResponse) authOrganizationsResponse {
	selectedOrg := strings.TrimSpace(me.SelectedOrganizationID)
	response := authOrganizationsResponse{
		PrincipalType: me.PrincipalType,
		SelectedOrg:   selectedOrg,
	}

	switch me.PrincipalType {
	case "api_key":
		response.Organizations = []authOrganizationSummary{{
			OrganizationID: me.OrganizationID,
			Selected:       true,
			Switchable:     false,
		}}
		response.APIKey = &authAPIKeySummary{
			APIKeyID:    me.APIKeyID,
			KeyType:     me.KeyType,
			Mode:        me.Mode,
			WorkspaceID: me.WorkspaceID,
		}
	case "user":
		response.Organizations = make([]authOrganizationSummary, 0, len(me.OrganizationMemberships))
		for _, membership := range me.OrganizationMemberships {
			response.Organizations = append(response.Organizations, authOrganizationSummary{
				OrganizationID: membership.OrganizationID,
				Role:           membership.Role,
				Selected:       selectedOrg != "" && membership.OrganizationID == selectedOrg,
				Switchable:     true,
			})
		}
	}

	return response
}

type authSnapshot struct {
	resolved       authconfig.Resolved
	session        *authstore.Session
	refreshToken   string
	refreshBackend string
	mode           authmode.Mode
	sessionMatch   bool
	config         authConfigDetails
	sources        whoAmISources
}

func loadAuthSnapshot(cmd *cli.Command) (authSnapshot, error) {
	resolved, err := authconfig.Resolve(cmd)
	if err != nil {
		return authSnapshot{}, authmode.WrapConfigError(err)
	}
	configDetails, err := loadAuthConfigDetails()
	if err != nil {
		return authSnapshot{}, authmode.WrapConfigError(err)
	}
	session, err := authstore.LoadSession()
	if err != nil {
		return authSnapshot{}, autherror.New("session_load_failed", "Failed to load OAuth session", err.Error())
	}
	preferredBackend := ""
	if session != nil {
		preferredBackend = session.StorageBackend
	}
	refreshToken, backend, err := authstore.LoadRefreshTokenWithPreferredBackend(preferredBackend)
	if err != nil {
		return authSnapshot{}, autherror.New("refresh_token_load_failed", "Failed to load refresh token", err.Error())
	}

	sessionMatch := session != nil && authmode.MatchError(resolved, session) == nil
	snapshot := authSnapshot{
		resolved:       resolved,
		session:        session,
		refreshToken:   refreshToken,
		refreshBackend: backend,
		mode:           authmode.DescribeMode(resolved, session),
		sessionMatch:   sessionMatch,
		config:         configDetails,
		sources: whoAmISources{
			APIKey:       resolved.Sources.APIKey,
			IssuerURL:    resolved.Sources.IssuerURL,
			ClientID:     resolved.Sources.ClientID,
			Scopes:       resolved.Sources.Scopes,
			Audience:     resolved.Sources.Audience,
			SelectedOrg:  resolved.Sources.SelectedOrg,
			Session:      authconfig.SourceUnset,
			RefreshToken: refreshTokenSource(refreshToken, backend),
		},
	}
	if session != nil {
		snapshot.sources.Session = authconfig.SourceSessionCache
	}
	return snapshot, nil
}

func loadAuthConfigDetails() (authConfigDetails, error) {
	path, err := authstore.ConfigFilePath()
	if err != nil {
		return authConfigDetails{}, err
	}
	details := authConfigDetails{Path: path}
	if _, statErr := os.Stat(path); statErr == nil {
		details.Present = true
	} else if !os.IsNotExist(statErr) {
		return authConfigDetails{}, statErr
	}
	return details, nil
}

func buildAuthStatusResponse(snapshot authSnapshot) (authStatusResponse, bool) {
	requestedScopes := append([]string(nil), snapshot.resolved.Scopes...)
	grantedScopes := append([]string(nil), snapshot.sessionGrantedScopes()...)
	missing := missingScopes(requestedScopes, grantedScopes)
	refreshTokenPresent := strings.TrimSpace(snapshot.refreshToken) != ""
	storedSession := buildStoredOAuthSession(snapshot, grantedScopes, missing, refreshTokenPresent)

	response := authStatusResponse{
		Authenticated:        false,
		EffectiveMode:        string(snapshot.mode),
		ActiveSource:         "none",
		Config:               snapshot.config,
		APIKeyConfigured:     strings.TrimSpace(snapshot.resolved.APIKey) != "",
		APIKeyOverridesOAuth: strings.TrimSpace(snapshot.resolved.APIKey) != "" && snapshot.session != nil,
		IssuerURL:            snapshot.resolved.IssuerURL,
		ClientID:             snapshot.resolved.ClientID,
		Audience:             snapshot.resolved.Audience,
		SelectedOrg:          snapshot.resolved.SelectedOrg,
		RequestedScopes:      requestedScopes,
		SessionPresent:       snapshot.session != nil,
		SessionMatch:         snapshot.sessionMatch,
		RefreshTokenPresent:  refreshTokenPresent,
		Sources:              snapshot.sources,
	}
	applyOAuthConfigWarnings(&response, snapshot)

	switch {
	case strings.TrimSpace(snapshot.resolved.APIKey) != "":
		response.Authenticated = true
		response.ActiveSource = "api_key"
		response.StoredOAuthSession = storedSession
		if snapshot.session != nil {
			response.Warnings = mergeStrings(response.Warnings, []string{
				"API-key mode is overriding the stored OAuth session.",
			})
			response.Actions = mergeStrings(response.Actions, []string{
				"Clear `--api-key` or `BOLTZ_API_KEY` to activate the stored OAuth session.",
			})
		}
	case snapshot.sessionMatch:
		response.ActiveSource = "oauth_session"
		response.Authenticated = oauthSessionUsable(snapshot.session, snapshot.refreshToken)
		applyStoredOAuthSessionToActiveResponse(&response, storedSession)
		if len(missing) > 0 {
			response.Warnings = mergeStrings(response.Warnings, []string{authStatusWarningScopeMismatch})
			response.Actions = mergeStrings(response.Actions, []string{
				"Run `boltz-api auth login` again if you need the full configured scope set.",
			})
		}
		if !response.Authenticated {
			response.Actions = mergeStrings(response.Actions, []string{
				"Run `boltz-api auth login` again to create a usable OAuth session.",
				"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode instead.",
			})
		}
	case snapshot.session != nil:
		response.StoredOAuthSession = storedSession
		response.Warnings = mergeStrings(response.Warnings, []string{
			"Stored OAuth session does not match the current auth settings.",
		})
		response.Actions = mergeStrings(response.Actions, []string{
			fmt.Sprintf("Run `boltz-api auth login` again for issuer %q and client %q.", snapshot.resolved.IssuerURL, snapshot.resolved.ClientID),
			"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode instead.",
		})
	default:
		response.Actions = mergeStrings(response.Actions, []string{
			"Run `boltz-api auth login` to create an OAuth session.",
			"Set `--api-key` or `BOLTZ_API_KEY` to use API-key mode.",
		})
	}

	return response, response.Authenticated
}

func applyOAuthConfigWarnings(response *authStatusResponse, snapshot authSnapshot) {
	if response == nil {
		return
	}

	if looksLikePlaceholderIssuer(snapshot.resolved.IssuerURL) {
		response.Warnings = mergeStrings(response.Warnings, []string{
			fmt.Sprintf("OAuth issuer URL looks like a placeholder: %q.", snapshot.resolved.IssuerURL),
		})
		response.Actions = mergeStrings(response.Actions, []string{
			oauthConfigAction(snapshot),
		})
	}

	if looksLikePlaceholderClientID(snapshot.resolved.ClientID) {
		response.Warnings = mergeStrings(response.Warnings, []string{
			fmt.Sprintf("OAuth client ID looks like a placeholder: %q.", snapshot.resolved.ClientID),
		})
		response.Actions = mergeStrings(response.Actions, []string{
			oauthConfigAction(snapshot),
		})
	}
}

func oauthConfigAction(snapshot authSnapshot) string {
	if snapshot.config.Path != "" && snapshot.config.Present {
		return fmt.Sprintf("Inspect %s or override OAuth settings with `--auth-issuer-url` and `--auth-client-id`.", snapshot.config.Path)
	}
	return "Override OAuth settings with `--auth-issuer-url` and `--auth-client-id`, or remove stale environment values."
}

func looksLikePlaceholderIssuer(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return false
	}
	return strings.Contains(value, "issuer.example.") ||
		strings.Contains(value, "example.com") ||
		strings.Contains(value, "example.org") ||
		strings.Contains(value, "example.net") ||
		strings.Contains(value, "placeholder")
}

func looksLikePlaceholderClientID(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return false
	}
	return value == "client-123" ||
		strings.Contains(value, "example") ||
		strings.Contains(value, "placeholder")
}

func buildStoredOAuthSession(snapshot authSnapshot, grantedScopes []string, missing []string, refreshTokenPresent bool) *oauthSessionDetails {
	if snapshot.session == nil {
		return nil
	}

	details := &oauthSessionDetails{
		GrantedScopes:       append([]string(nil), grantedScopes...),
		GrantedScopesKnown:  len(grantedScopes) > 0,
		MissingScopes:       append([]string(nil), missing...),
		RefreshTokenPresent: refreshTokenPresent,
		Refreshable:         snapshot.sessionMatch && refreshTokenPresent,
		StorageBackend:      snapshot.session.StorageBackend,
		Identity:            summarizeIdentity(snapshot.session.Identity),
	}
	if details.StorageBackend == "" {
		details.StorageBackend = snapshot.refreshBackend
	}
	if !snapshot.session.Expiry.IsZero() {
		expiry := snapshot.session.Expiry
		details.Expiry = &expiry
	}
	return details
}

func applyStoredOAuthSessionToActiveResponse(response *authStatusResponse, details *oauthSessionDetails) {
	if response == nil || details == nil {
		return
	}
	response.GrantedScopes = append([]string(nil), details.GrantedScopes...)
	response.GrantedScopesKnown = details.GrantedScopesKnown
	response.MissingScopes = append([]string(nil), details.MissingScopes...)
	response.Refreshable = details.Refreshable
	response.StorageBackend = details.StorageBackend
	response.Expiry = details.Expiry
	response.Identity = details.Identity
}

func applyValidateModeNotes(response *authValidateResponse, mode authmode.Mode) {
	if response == nil {
		return
	}
	if mode != authmode.ModeAPIKey {
		return
	}
	response.Warnings = mergeStrings(response.Warnings, []string{
		"API-key validation is local only; this command confirms that an API key is configured.",
	})
	response.Actions = mergeStrings(response.Actions, []string{
		"Run a real API request to verify the API key against the server.",
	})
}

func restoreLoginState(previousSession *authstore.Session, previousRefreshToken string) error {
	var restoreErrs []string
	if previousSession != nil {
		if err := authstore.SaveSession(*previousSession); err != nil {
			restoreErrs = append(restoreErrs, "session: "+err.Error())
		}
	} else if err := authstore.ClearSession(); err != nil {
		restoreErrs = append(restoreErrs, "session: "+err.Error())
	}

	if strings.TrimSpace(previousRefreshToken) != "" {
		if _, err := authstore.SaveRefreshToken(previousRefreshToken); err != nil {
			restoreErrs = append(restoreErrs, "refresh token: "+err.Error())
		}
	} else if err := authstore.ClearRefreshToken(); err != nil {
		restoreErrs = append(restoreErrs, "refresh token: "+err.Error())
	}

	if len(restoreErrs) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(restoreErrs, "; "))
}

func (s authSnapshot) sessionGrantedScopes() []string {
	if s.session == nil {
		return nil
	}
	return s.session.GrantedScopes
}

func summarizeIdentity(identity authstore.Identity) *whoAmIIdentity {
	if identity.Subject == "" && identity.Email == "" && identity.Name == "" && identity.PreferredUsername == "" {
		return nil
	}
	return &whoAmIIdentity{
		Subject:           identity.Subject,
		Email:             identity.Email,
		Name:              identity.Name,
		PreferredUsername: identity.PreferredUsername,
	}
}

func refreshTokenSource(token string, backend string) authconfig.Source {
	if strings.TrimSpace(token) == "" {
		return authconfig.SourceUnset
	}
	switch backend {
	case "keyring":
		return authconfig.SourceKeyring
	case "file":
		return authconfig.SourceFile
	default:
		return authconfig.SourceUnset
	}
}

func missingScopes(requested []string, granted []string) []string {
	if len(requested) == 0 || len(granted) == 0 {
		return nil
	}

	grantedSet := make(map[string]struct{}, len(granted))
	for _, scope := range granted {
		grantedSet[strings.TrimSpace(scope)] = struct{}{}
	}

	missing := make([]string, 0, len(requested))
	for _, scope := range requested {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := grantedSet[scope]; ok {
			continue
		}
		missing = append(missing, scope)
	}
	if len(missing) == 0 {
		return nil
	}
	return missing
}

func oauthSessionUsable(session *authstore.Session, refreshToken string) bool {
	if session == nil {
		return false
	}
	if strings.TrimSpace(session.AccessToken) == "" {
		return strings.TrimSpace(refreshToken) != ""
	}
	if session.Expiry.IsZero() {
		return true
	}
	if session.Expiry.After(authNow().UTC().Add(60 * time.Second)) {
		return true
	}
	return strings.TrimSpace(refreshToken) != ""
}

func wasSessionRefreshed(before *authstore.Session, after *authstore.Session) bool {
	switch {
	case before == nil || after == nil:
		return false
	case strings.TrimSpace(before.AccessToken) != strings.TrimSpace(after.AccessToken):
		return true
	case !before.Expiry.Equal(after.Expiry):
		return true
	case !slices.Equal(before.GrantedScopes, after.GrantedScopes):
		return true
	default:
		return false
	}
}

func mergeStrings(existing []string, values []string) []string {
	if len(values) == 0 {
		return existing
	}
	seen := make(map[string]struct{}, len(existing)+len(values))
	merged := make([]string, 0, len(existing)+len(values))
	for _, value := range existing {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		merged = append(merged, value)
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		merged = append(merged, value)
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func showJSONValue(cmd *cli.Command, value any, title string) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return ShowJSON(gjson.ParseBytes(body), ShowJSONOpts{
		ExplicitFormat: cmd.Root().IsSet("format"),
		Format:         cmd.Root().String("format"),
		RawOutput:      cmd.Root().Bool("raw-output"),
		Stdout:         commandFileWriter(cmd),
		Stderr:         commandErrorWriter(cmd),
		Title:          title,
		Transform:      cmd.Root().String("transform"),
	})
}

func commandWriter(cmd *cli.Command) io.Writer {
	if root := cmd.Root(); root != nil && root.Writer != nil {
		return root.Writer
	}
	if cmd.Writer != nil {
		return cmd.Writer
	}
	return os.Stdout
}

func commandErrorWriter(cmd *cli.Command) io.Writer {
	if root := cmd.Root(); root != nil && root.ErrWriter != nil {
		return root.ErrWriter
	}
	if cmd.ErrWriter != nil {
		return cmd.ErrWriter
	}
	return os.Stderr
}

func commandFileWriter(cmd *cli.Command) *os.File {
	if root := cmd.Root(); root != nil {
		if file, ok := root.Writer.(*os.File); ok {
			return file
		}
	}
	if file, ok := cmd.Writer.(*os.File); ok {
		return file
	}
	return nil
}
