package oauthclient

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	callbackPath          = "/oauth/callback"
	callbackTimeout       = 15 * time.Minute
	deviceCodeGrantType   = "urn:ietf:params:oauth:grant-type:device_code"
	devicePollingMaxDelay = 30 * time.Second
	httpTimeout           = 15 * time.Second
)

var openBrowser = defaultOpenBrowser

type Config struct {
	IssuerURL        string
	ClientID         string
	Scopes           []string
	Audience         string
	AuthorizationURL string
	TokenURL         string
	UserInfoURL      string
	RevocationURL    string
	ListenPort       int
	HTTPClient       *http.Client
	Output           io.Writer
	OnDeviceCode     func(DeviceAuthorization)
}

type ProviderMetadata struct {
	IssuerURL              string
	ComputeAPIBaseURL      string
	AuthorizationURL       string
	DeviceAuthorizationURL string
	TokenURL               string
	UserInfoURL            string
	RevocationURL          string
	JWKSURL                string
	Algorithms             []string
}

type TokenSet struct {
	AccessToken   string
	RefreshToken  string
	TokenType     string
	Expiry        time.Time
	IDToken       string
	GrantedScopes []string
}

type LoginResult struct {
	Provider         ProviderMetadata
	Tokens           TokenSet
	Identity         authstore.Identity
	AuthorizationURL string
}

type DeviceAuthorization struct {
	VerificationURI         string
	VerificationURIComplete string
	UserCode                string
	ExpiresIn               int64
	Interval                int64
}

type RefreshConfig struct {
	Provider     ProviderMetadata
	ClientID     string
	RefreshToken string
	Resource     string
	HTTPClient   *http.Client
}

type RevokeConfig struct {
	Endpoint      string
	ClientID      string
	Token         string
	TokenTypeHint string
	HTTPClient    *http.Client
}

type discoveryDocument struct {
	IssuerURL              string   `json:"issuer"`
	ComputeAPIBaseURL      string   `json:"boltz_compute_api_base_url"`
	AuthorizationURL       string   `json:"authorization_endpoint"`
	DeviceAuthorizationURL string   `json:"device_authorization_endpoint"`
	TokenURL               string   `json:"token_endpoint"`
	UserInfoURL            string   `json:"userinfo_endpoint"`
	RevocationURL          string   `json:"revocation_endpoint"`
	JWKSURL                string   `json:"jwks_uri"`
	Algorithms             []string `json:"id_token_signing_alg_values_supported"`
}

type callbackResult struct {
	code string
	err  error
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	IDToken          string `json:"id_token"`
	Scope            string `json:"scope"`
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type deviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int64  `json:"expires_in"`
	Interval                int64  `json:"interval"`
	ErrorCode               string `json:"error"`
	ErrorDescription        string `json:"error_description"`
}

type TokenError struct {
	StatusCode       int
	ErrorCode        string
	ErrorDescription string
}

func (e *TokenError) Error() string {
	if e == nil {
		return ""
	}
	if e.ErrorCode == "" {
		return fmt.Sprintf("token endpoint returned HTTP %d", e.StatusCode)
	}
	if e.ErrorDescription == "" {
		return fmt.Sprintf("token endpoint returned %s", e.ErrorCode)
	}
	return fmt.Sprintf("token endpoint returned %s: %s", e.ErrorCode, e.ErrorDescription)
}

func Login(ctx context.Context, cfg Config) (*LoginResult, error) {
	if strings.TrimSpace(cfg.IssuerURL) == "" {
		return nil, fmt.Errorf("issuer URL is required")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	httpClient := resolveHTTPClient(cfg.HTTPClient)
	ctx = oidc.ClientContext(ctx, httpClient)

	providerMetadata, provider, err := discoverProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(providerMetadata.AuthorizationURL) == "" || strings.TrimSpace(providerMetadata.TokenURL) == "" {
		return nil, fmt.Errorf("provider discovery did not return authorization and token endpoints")
	}

	listener, redirectURL, err := listenLoopback(cfg.ListenPort)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	verifier := oauth2.GenerateVerifier()
	state, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	nonce, err := randomToken(24)
	if err != nil {
		return nil, err
	}

	oauthConfig := oauth2.Config{
		ClientID:    cfg.ClientID,
		Endpoint:    oauth2.Endpoint{AuthURL: providerMetadata.AuthorizationURL, TokenURL: providerMetadata.TokenURL},
		RedirectURL: redirectURL,
		Scopes:      append([]string(nil), cfg.Scopes...),
	}

	authOptions := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
		oidc.Nonce(nonce),
	}
	if strings.TrimSpace(cfg.Audience) != "" {
		authOptions = append(authOptions, oauth2.SetAuthURLParam("audience", cfg.Audience))
		authOptions = append(authOptions, oauth2.SetAuthURLParam("resource", cfg.Audience))
	}

	authURL := oauthConfig.AuthCodeURL(state, authOptions...)
	if cfg.Output != nil {
		fmt.Fprintf(cfg.Output, "Open this URL to authenticate:\n%s\n", authURL)
	}
	if err := openBrowser(authURL); err != nil && cfg.Output != nil {
		fmt.Fprintf(cfg.Output, "Could not open browser automatically: %v\n", err)
	}

	code, err := waitForCallback(ctx, listener, state)
	if err != nil {
		return nil, err
	}

	exchangeOptions := []oauth2.AuthCodeOption{
		oauth2.VerifierOption(verifier),
		oauth2.SetAuthURLParam("client_id", cfg.ClientID),
	}
	if strings.TrimSpace(cfg.Audience) != "" {
		exchangeOptions = append(exchangeOptions, oauth2.SetAuthURLParam("resource", cfg.Audience))
	}

	token, err := oauthConfig.Exchange(ctx, code, exchangeOptions...)
	if err != nil {
		return nil, err
	}

	tokens := extractTokenSet(token)
	var idToken *oidc.IDToken
	if tokens.IDToken != "" {
		idToken, err = verifyIDToken(ctx, provider, cfg.ClientID, tokens.IDToken)
		if err != nil {
			return nil, err
		}
		if idToken.Nonce != nonce {
			return nil, fmt.Errorf("received ID token with unexpected nonce")
		}
	}

	identity, err := resolveIdentity(ctx, provider, tokens.AccessToken, idToken)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Provider:         providerMetadata,
		Tokens:           tokens,
		Identity:         identity,
		AuthorizationURL: authURL,
	}, nil
}

func DeviceLogin(ctx context.Context, cfg Config) (*LoginResult, error) {
	if strings.TrimSpace(cfg.IssuerURL) == "" {
		return nil, fmt.Errorf("issuer URL is required")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("client ID is required")
	}

	httpClient := resolveHTTPClient(cfg.HTTPClient)
	ctx = oidc.ClientContext(ctx, httpClient)

	providerMetadata, provider, err := discoverProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(providerMetadata.DeviceAuthorizationURL) == "" {
		return nil, fmt.Errorf("provider discovery did not return a device authorization endpoint")
	}
	if strings.TrimSpace(providerMetadata.TokenURL) == "" {
		return nil, fmt.Errorf("provider discovery did not return a token endpoint")
	}

	device, err := requestDeviceAuthorization(ctx, httpClient, providerMetadata.DeviceAuthorizationURL, cfg)
	if err != nil {
		return nil, err
	}

	if cfg.OnDeviceCode != nil {
		cfg.OnDeviceCode(DeviceAuthorization{
			VerificationURI:         strings.TrimSpace(device.VerificationURI),
			VerificationURIComplete: strings.TrimSpace(device.VerificationURIComplete),
			UserCode:                strings.TrimSpace(device.UserCode),
			ExpiresIn:               device.ExpiresIn,
			Interval:                device.Interval,
		})
	}

	verificationURL := strings.TrimSpace(device.VerificationURIComplete)
	if verificationURL == "" {
		verificationURL = strings.TrimSpace(device.VerificationURI)
	}
	if cfg.Output != nil {
		fmt.Fprintf(cfg.Output, "Open this URL to authenticate:\n%s\n", verificationURL)
		if strings.TrimSpace(device.UserCode) != "" {
			fmt.Fprintf(cfg.Output, "Enter this code:\n%s\n", device.UserCode)
		}
	}
	if verificationURL != "" {
		if err := openBrowser(verificationURL); err != nil && cfg.Output != nil {
			fmt.Fprintf(cfg.Output, "Could not open browser automatically: %v\n", err)
		}
	}

	tokens, err := pollDeviceToken(ctx, httpClient, providerMetadata.TokenURL, cfg, device)
	if err != nil {
		return nil, err
	}

	var idToken *oidc.IDToken
	if tokens.IDToken != "" {
		idToken, err = verifyIDToken(ctx, provider, cfg.ClientID, tokens.IDToken)
		if err != nil {
			return nil, err
		}
	}

	identity, err := resolveIdentity(ctx, provider, tokens.AccessToken, idToken)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Provider:         providerMetadata,
		Tokens:           *tokens,
		Identity:         identity,
		AuthorizationURL: verificationURL,
	}, nil
}

func ResolveProviderMetadata(ctx context.Context, cfg Config) (ProviderMetadata, error) {
	httpClient := resolveHTTPClient(cfg.HTTPClient)
	ctx = oidc.ClientContext(ctx, httpClient)

	metadata, _, err := discoverProvider(ctx, cfg)
	if err != nil {
		return ProviderMetadata{}, err
	}
	return metadata, nil
}

func Refresh(ctx context.Context, cfg RefreshConfig) (*TokenSet, error) {
	if strings.TrimSpace(cfg.Provider.TokenURL) == "" {
		return nil, fmt.Errorf("token endpoint is required for refresh")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("client ID is required for refresh")
	}
	if strings.TrimSpace(cfg.RefreshToken) == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", cfg.RefreshToken)
	form.Set("client_id", cfg.ClientID)
	if strings.TrimSpace(cfg.Resource) != "" {
		form.Set("resource", strings.TrimSpace(cfg.Resource))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Provider.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := resolveHTTPClient(cfg.HTTPClient).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decoded tokenResponse
	if len(body) > 0 {
		_ = json.Unmarshal(body, &decoded)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &TokenError{
			StatusCode:       resp.StatusCode,
			ErrorCode:        decoded.ErrorCode,
			ErrorDescription: decoded.ErrorDescription,
		}
	}

	tokens := TokenSet{
		AccessToken:   strings.TrimSpace(decoded.AccessToken),
		RefreshToken:  strings.TrimSpace(decoded.RefreshToken),
		TokenType:     strings.TrimSpace(decoded.TokenType),
		IDToken:       strings.TrimSpace(decoded.IDToken),
		GrantedScopes: parseScopes(decoded.Scope),
	}
	if decoded.ExpiresIn > 0 {
		tokens.Expiry = time.Now().UTC().Add(time.Duration(decoded.ExpiresIn) * time.Second)
	}

	if tokens.IDToken != "" {
		providerMetadata := cfg.Provider
		if (strings.TrimSpace(providerMetadata.JWKSURL) == "" || len(providerMetadata.Algorithms) == 0) && strings.TrimSpace(providerMetadata.IssuerURL) != "" {
			discovered, err := ResolveProviderMetadata(ctx, Config{
				IssuerURL:        providerMetadata.IssuerURL,
				AuthorizationURL: providerMetadata.AuthorizationURL,
				TokenURL:         providerMetadata.TokenURL,
				UserInfoURL:      providerMetadata.UserInfoURL,
				RevocationURL:    providerMetadata.RevocationURL,
				HTTPClient:       cfg.HTTPClient,
			})
			if err == nil {
				if strings.TrimSpace(providerMetadata.JWKSURL) == "" {
					providerMetadata.JWKSURL = discovered.JWKSURL
				}
				if len(providerMetadata.Algorithms) == 0 {
					providerMetadata.Algorithms = append([]string(nil), discovered.Algorithms...)
				}
			}
		}

		verifyCtx := oidc.ClientContext(ctx, resolveHTTPClient(cfg.HTTPClient))
		provider := providerFromMetadata(verifyCtx, providerMetadata)
		if _, err := verifyIDToken(verifyCtx, provider, cfg.ClientID, tokens.IDToken); err != nil {
			return nil, err
		}
	}

	return &tokens, nil
}

func Revoke(ctx context.Context, cfg RevokeConfig) error {
	if strings.TrimSpace(cfg.Endpoint) == "" || strings.TrimSpace(cfg.Token) == "" {
		return nil
	}

	form := url.Values{}
	form.Set("token", cfg.Token)
	if strings.TrimSpace(cfg.ClientID) != "" {
		form.Set("client_id", cfg.ClientID)
	}
	if strings.TrimSpace(cfg.TokenTypeHint) != "" {
		form.Set("token_type_hint", cfg.TokenTypeHint)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := resolveHTTPClient(cfg.HTTPClient).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("revocation endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func IsInvalidGrant(err error) bool {
	var tokenErr *TokenError
	if errors.As(err, &tokenErr) {
		return tokenErr.ErrorCode == "invalid_grant"
	}
	return strings.Contains(err.Error(), "invalid_grant")
}

func discoverProvider(ctx context.Context, cfg Config) (ProviderMetadata, *oidc.Provider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return ProviderMetadata{}, nil, err
	}

	endpoint := provider.Endpoint()
	var claims discoveryDocument
	if err := provider.Claims(&claims); err != nil {
		return ProviderMetadata{}, nil, err
	}

	metadata := ProviderMetadata{
		IssuerURL:              strings.TrimSpace(claims.IssuerURL),
		ComputeAPIBaseURL:      strings.TrimSpace(claims.ComputeAPIBaseURL),
		AuthorizationURL:       strings.TrimSpace(endpoint.AuthURL),
		DeviceAuthorizationURL: strings.TrimSpace(claims.DeviceAuthorizationURL),
		TokenURL:               strings.TrimSpace(endpoint.TokenURL),
		UserInfoURL:            strings.TrimSpace(provider.UserInfoEndpoint()),
		RevocationURL:          strings.TrimSpace(claims.RevocationURL),
		JWKSURL:                strings.TrimSpace(claims.JWKSURL),
		Algorithms:             append([]string(nil), claims.Algorithms...),
	}
	if metadata.IssuerURL == "" {
		metadata.IssuerURL = strings.TrimSpace(cfg.IssuerURL)
	}
	if strings.TrimSpace(cfg.AuthorizationURL) != "" {
		metadata.AuthorizationURL = strings.TrimSpace(cfg.AuthorizationURL)
	}
	if strings.TrimSpace(cfg.TokenURL) != "" {
		metadata.TokenURL = strings.TrimSpace(cfg.TokenURL)
	}
	if strings.TrimSpace(cfg.UserInfoURL) != "" {
		metadata.UserInfoURL = strings.TrimSpace(cfg.UserInfoURL)
	}
	if strings.TrimSpace(cfg.RevocationURL) != "" {
		metadata.RevocationURL = strings.TrimSpace(cfg.RevocationURL)
	}

	provider = providerFromMetadata(ctx, metadata)
	return metadata, provider, nil
}

func requestDeviceAuthorization(ctx context.Context, client *http.Client, endpoint string, cfg Config) (*deviceAuthorizationResponse, error) {
	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	if len(cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	if strings.TrimSpace(cfg.Audience) != "" {
		form.Set("resource", strings.TrimSpace(cfg.Audience))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decoded deviceAuthorizationResponse
	if len(body) > 0 {
		_ = json.Unmarshal(body, &decoded)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &TokenError{
			StatusCode:       resp.StatusCode,
			ErrorCode:        decoded.ErrorCode,
			ErrorDescription: decoded.ErrorDescription,
		}
	}
	if strings.TrimSpace(decoded.DeviceCode) == "" {
		return nil, fmt.Errorf("device authorization response did not include a device code")
	}
	if strings.TrimSpace(decoded.VerificationURI) == "" && strings.TrimSpace(decoded.VerificationURIComplete) == "" {
		return nil, fmt.Errorf("device authorization response did not include a verification URL")
	}
	return &decoded, nil
}

func pollDeviceToken(ctx context.Context, client *http.Client, endpoint string, cfg Config, device *deviceAuthorizationResponse) (*TokenSet, error) {
	interval := time.Duration(device.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	expiresIn := time.Duration(device.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = callbackTimeout
	}

	pollCtx, cancel := context.WithTimeout(ctx, expiresIn)
	defer cancel()

	for {
		tokens, err := exchangeDeviceCode(pollCtx, client, endpoint, cfg, device.DeviceCode)
		if err == nil {
			return tokens, nil
		}

		var tokenErr *TokenError
		if !errors.As(err, &tokenErr) {
			return nil, err
		}

		switch tokenErr.ErrorCode {
		case "authorization_pending":
			// Expected while the user is still approving the code.
		case "slow_down":
			interval += time.Duration(device.Interval) * time.Second
			if interval > devicePollingMaxDelay {
				interval = devicePollingMaxDelay
			}
		default:
			return nil, err
		}

		select {
		case <-time.After(interval):
		case <-pollCtx.Done():
			return nil, fmt.Errorf("timed out waiting for device authorization")
		}
	}
}

func exchangeDeviceCode(ctx context.Context, client *http.Client, endpoint string, cfg Config, deviceCode string) (*TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", deviceCodeGrantType)
	form.Set("client_id", cfg.ClientID)
	form.Set("device_code", deviceCode)
	if strings.TrimSpace(cfg.Audience) != "" {
		form.Set("resource", strings.TrimSpace(cfg.Audience))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decoded tokenResponse
	if len(body) > 0 {
		_ = json.Unmarshal(body, &decoded)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &TokenError{
			StatusCode:       resp.StatusCode,
			ErrorCode:        decoded.ErrorCode,
			ErrorDescription: decoded.ErrorDescription,
		}
	}

	tokens := TokenSet{
		AccessToken:   strings.TrimSpace(decoded.AccessToken),
		RefreshToken:  strings.TrimSpace(decoded.RefreshToken),
		TokenType:     strings.TrimSpace(decoded.TokenType),
		IDToken:       strings.TrimSpace(decoded.IDToken),
		GrantedScopes: parseScopes(decoded.Scope),
	}
	if decoded.ExpiresIn > 0 {
		tokens.Expiry = time.Now().UTC().Add(time.Duration(decoded.ExpiresIn) * time.Second)
	}
	return &tokens, nil
}

func providerFromMetadata(ctx context.Context, metadata ProviderMetadata) *oidc.Provider {
	return (&oidc.ProviderConfig{
		IssuerURL:   metadata.IssuerURL,
		AuthURL:     metadata.AuthorizationURL,
		TokenURL:    metadata.TokenURL,
		UserInfoURL: metadata.UserInfoURL,
		JWKSURL:     metadata.JWKSURL,
		Algorithms:  append([]string(nil), metadata.Algorithms...),
	}).NewProvider(ctx)
}

func verifyIDToken(ctx context.Context, provider *oidc.Provider, clientID, rawIDToken string) (*oidc.IDToken, error) {
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return verifier.Verify(ctx, rawIDToken)
}

func resolveIdentity(ctx context.Context, provider *oidc.Provider, accessToken string, idToken *oidc.IDToken) (authstore.Identity, error) {
	if accessToken != "" && provider.UserInfoEndpoint() != "" {
		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}))
		if err == nil {
			var claims map[string]any
			if claimsErr := userInfo.Claims(&claims); claimsErr != nil {
				return authstore.Identity{}, claimsErr
			}
			return identityFromClaims(claims), nil
		}
	}

	if idToken != nil {
		var claims map[string]any
		if err := idToken.Claims(&claims); err != nil {
			return authstore.Identity{}, err
		}
		return identityFromClaims(claims), nil
	}

	return authstore.Identity{}, nil
}

func identityFromClaims(claims map[string]any) authstore.Identity {
	return authstore.Identity{
		Subject:           stringClaim(claims, "sub"),
		Email:             stringClaim(claims, "email"),
		Name:              stringClaim(claims, "name"),
		PreferredUsername: stringClaim(claims, "preferred_username"),
		Claims:            claims,
	}
}

func stringClaim(claims map[string]any, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func extractTokenSet(token *oauth2.Token) TokenSet {
	if token == nil {
		return TokenSet{}
	}
	idToken, _ := token.Extra("id_token").(string)
	grantedScopes := parseScopeValue(token.Extra("scope"))
	return TokenSet{
		AccessToken:   strings.TrimSpace(token.AccessToken),
		RefreshToken:  strings.TrimSpace(token.RefreshToken),
		TokenType:     strings.TrimSpace(token.TokenType),
		Expiry:        token.Expiry.UTC(),
		IDToken:       strings.TrimSpace(idToken),
		GrantedScopes: grantedScopes,
	}
}

func parseScopeValue(raw any) []string {
	switch typed := raw.(type) {
	case string:
		return parseScopes(typed)
	case []string:
		return parseScopes(strings.Join(typed, " "))
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if scope, ok := item.(string); ok {
				parts = append(parts, scope)
			}
		}
		return parseScopes(strings.Join(parts, " "))
	default:
		return nil
	}
}

func parseScopes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	normalized := strings.ReplaceAll(raw, ",", " ")
	parts := strings.Fields(normalized)
	if len(parts) == 0 {
		return nil
	}

	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		scope := strings.TrimSpace(part)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		result = append(result, scope)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func listenLoopback(port int) (net.Listener, string, error) {
	address := "127.0.0.1:0"
	if port > 0 {
		address = fmt.Sprintf("127.0.0.1:%d", port)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, "", err
	}

	redirectURL := url.URL{
		Scheme: "http",
		Host:   listener.Addr().String(),
		Path:   callbackPath,
	}
	return listener, redirectURL.String(), nil
}

func waitForCallback(ctx context.Context, listener net.Listener, expectedState string) (string, error) {
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch {
		case query.Get("error") != "":
			http.Error(w, "Authentication failed. You can close this window.", http.StatusBadRequest)
			resultCh <- callbackResult{err: fmt.Errorf("authorization failed: %s", query.Get("error"))}
		case query.Get("state") != expectedState:
			http.Error(w, "Authentication failed. You can close this window.", http.StatusBadRequest)
			resultCh <- callbackResult{err: fmt.Errorf("received callback with invalid state")}
		case strings.TrimSpace(query.Get("code")) == "":
			http.Error(w, "Authentication failed. You can close this window.", http.StatusBadRequest)
			resultCh <- callbackResult{err: fmt.Errorf("authorization code missing from callback")}
		default:
			fmt.Fprintln(w, "Authentication complete. You can close this window.")
			resultCh <- callbackResult{code: query.Get("code")}
		}
	})

	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Shutdown(context.Background())

	waitCtx, cancel := context.WithTimeout(ctx, callbackTimeout)
	defer cancel()

	select {
	case result := <-resultCh:
		return result.code, result.err
	case <-waitCtx.Done():
		if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf(
				"timed out waiting for authorization callback after %s; the browser callback listener on %s has closed. Run `boltz-api auth login` again, or use `boltz-api auth login --device-code` if localhost callbacks are unreliable",
				callbackTimeout,
				listener.Addr().String(),
			)
		}
		return "", waitCtx.Err()
	}
}

func randomToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func resolveHTTPClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: httpTimeout}
}

func defaultOpenBrowser(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}
