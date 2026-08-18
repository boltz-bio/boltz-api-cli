package oauthclient

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
)

func TestLoginUsesIDTokenClaimsWhenUserInfoIsUnavailable(t *testing.T) {
	provider := newMockOIDCProvider(t, false)
	defer provider.Close()
	provider.SetComputeAPIBaseURL("https://api.browser.customer.example.com:8443")

	originalOpenBrowser := openBrowser
	openBrowser = func(target string) error {
		go visitURL(target)
		return nil
	}
	t.Cleanup(func() { openBrowser = originalOpenBrowser })

	var output bytes.Buffer
	result, err := Login(context.Background(), Config{
		IssuerURL: provider.IssuerURL(),
		ClientID:  "client-123",
		Scopes:    []string{"openid", "profile", "email"},
		Output:    &output,
	})
	require.NoError(t, err)
	require.Equal(t, "login-access", result.Tokens.AccessToken)
	require.Equal(t, "login-refresh", result.Tokens.RefreshToken)
	require.Equal(t, []string{"openid", "email"}, result.Tokens.GrantedScopes)
	require.Equal(t, "user@example.com", result.Identity.Email)
	require.Equal(t, "Test User", result.Identity.Name)
	require.Equal(t, "https://api.browser.customer.example.com:8443", result.Provider.ComputeAPIBaseURL)
	require.Contains(t, output.String(), provider.IssuerURL())
}

func TestLoginHonorsAuthorizationOverrideAndKeepsWorkingIfBrowserOpenFails(t *testing.T) {
	provider := newMockOIDCProvider(t, true)
	defer provider.Close()

	overrideURL := provider.IssuerURL() + "/authorize-override"

	originalOpenBrowser := openBrowser
	openBrowser = func(target string) error {
		go visitURL(target)
		return errors.New("browser unavailable")
	}
	t.Cleanup(func() { openBrowser = originalOpenBrowser })

	var output bytes.Buffer
	result, err := Login(context.Background(), Config{
		IssuerURL:        provider.IssuerURL(),
		ClientID:         "client-123",
		Scopes:           []string{"openid", "profile", "email"},
		Audience:         "boltz-compute-api",
		AuthorizationURL: overrideURL,
		Output:           &output,
	})
	require.NoError(t, err)
	require.Equal(t, "login-access", result.Tokens.AccessToken)
	require.Equal(t, "/authorize-override", provider.LastAuthorizationPath())
	require.Equal(t, "boltz-compute-api", provider.LastAuthorizationQuery().Get("audience"))
	require.Equal(t, "boltz-compute-api", provider.LastAuthorizationQuery().Get("resource"))
	require.Equal(t, "boltz-compute-api", provider.LastTokenForm().Get("resource"))
	require.Contains(t, output.String(), overrideURL)
	require.Contains(t, output.String(), "Could not open browser automatically")
	require.Equal(t, "userinfo@example.com", result.Identity.Email)
}

func TestWaitForCallbackTimeoutGivesActionableHint(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	_, err = waitForCallback(ctx, listener, "expected-state")
	require.Error(t, err)
	require.Contains(t, err.Error(), "timed out waiting for authorization callback after 15m0s")
	require.Contains(t, err.Error(), listener.Addr().String())
	require.Contains(t, err.Error(), "boltz-api auth login --device-code")
}

func TestDeviceLoginPrintsCodeAndPollsTokenEndpoint(t *testing.T) {
	provider := newMockOIDCProvider(t, true)
	defer provider.Close()
	provider.SetComputeAPIBaseURL("http://127.0.0.1:9443")

	originalOpenBrowser := openBrowser
	openBrowser = func(target string) error {
		require.Equal(t, provider.IssuerURL()+"/device?user_code=WDJB-MJHT", target)
		return nil
	}
	t.Cleanup(func() { openBrowser = originalOpenBrowser })

	var output bytes.Buffer
	result, err := DeviceLogin(context.Background(), Config{
		IssuerURL: provider.IssuerURL(),
		ClientID:  "client-123",
		Scopes:    []string{"openid", "profile", "email", "offline_access", "compute:run"},
		Audience:  "boltz-compute-api",
		Output:    &output,
	})
	require.NoError(t, err)
	require.Equal(t, "login-access", result.Tokens.AccessToken)
	require.Equal(t, "login-refresh", result.Tokens.RefreshToken)
	require.Equal(t, []string{"openid", "email"}, result.Tokens.GrantedScopes)
	require.Equal(t, "userinfo@example.com", result.Identity.Email)
	require.Equal(t, "http://127.0.0.1:9443", result.Provider.ComputeAPIBaseURL)
	require.Contains(t, output.String(), provider.IssuerURL()+"/device?user_code=WDJB-MJHT")
	require.Contains(t, output.String(), "WDJB-MJHT")
	require.Equal(t, "client-123", provider.LastDeviceAuthorizationForm().Get("client_id"))
	require.Equal(t, "boltz-compute-api", provider.LastDeviceAuthorizationForm().Get("resource"))
	require.Equal(t, deviceCodeGrantType, provider.LastTokenForm().Get("grant_type"))
	require.Equal(t, "device-code-123", provider.LastTokenForm().Get("device_code"))
	require.Equal(t, "boltz-compute-api", provider.LastTokenForm().Get("resource"))
}

func TestResolveProviderMetadataOptionalComputeAPIBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{name: "absent"},
		{
			name:     "custom domain and port",
			baseURL:  "https://api.customer.example.com:8443",
			expected: "https://api.customer.example.com:8443",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newMockOIDCProvider(t, true)
			defer provider.Close()
			provider.SetComputeAPIBaseURL(test.baseURL)

			metadata, err := ResolveProviderMetadata(context.Background(), Config{IssuerURL: provider.IssuerURL()})
			require.NoError(t, err)
			require.Equal(t, test.expected, metadata.ComputeAPIBaseURL)
		})
	}
}

func TestRefreshReturnsRotatedTokens(t *testing.T) {
	provider := newMockOIDCProvider(t, true)
	defer provider.Close()

	tokens, err := Refresh(context.Background(), RefreshConfig{
		Provider: ProviderMetadata{
			IssuerURL:  provider.IssuerURL(),
			TokenURL:   provider.IssuerURL() + "/token",
			JWKSURL:    provider.IssuerURL() + "/jwks",
			Algorithms: []string{"RS256"},
		},
		ClientID:     "client-123",
		RefreshToken: "refresh-token",
		Resource:     "boltz-compute-api",
	})
	require.NoError(t, err)
	require.Equal(t, "refreshed-access", tokens.AccessToken)
	require.Equal(t, "rotated-refresh", tokens.RefreshToken)
	require.Equal(t, "Bearer", tokens.TokenType)
	require.Equal(t, []string{"openid", "email"}, tokens.GrantedScopes)
	require.Equal(t, "boltz-compute-api", provider.LastTokenForm().Get("resource"))
}

func TestRefreshDiscoversSigningMetadataWhenNotProvided(t *testing.T) {
	provider := newMockOIDCProvider(t, true)
	defer provider.Close()

	tokens, err := Refresh(context.Background(), RefreshConfig{
		Provider: ProviderMetadata{
			IssuerURL: provider.IssuerURL(),
			TokenURL:  provider.IssuerURL() + "/token",
		},
		ClientID:     "client-123",
		RefreshToken: "refresh-token",
	})
	require.NoError(t, err)
	require.Equal(t, "refreshed-access", tokens.AccessToken)
	require.Equal(t, "rotated-refresh", tokens.RefreshToken)
}

func TestRevokePostsTokenToEndpoint(t *testing.T) {
	provider := newMockOIDCProvider(t, true)
	defer provider.Close()

	require.NoError(t, Revoke(context.Background(), RevokeConfig{
		Endpoint:      provider.IssuerURL() + "/revoke",
		ClientID:      "client-123",
		Token:         "refresh-token",
		TokenTypeHint: "refresh_token",
	}))

	values := provider.LastRevocation()
	require.Equal(t, "refresh-token", values.Get("token"))
	require.Equal(t, "client-123", values.Get("client_id"))
	require.Equal(t, "refresh_token", values.Get("token_type_hint"))
}

type mockOIDCProvider struct {
	t                 *testing.T
	server            *httptest.Server
	signer            jose.Signer
	publicJWKSetJSON  []byte
	includeUserInfo   bool
	computeAPIBaseURL string

	mu                     sync.Mutex
	lastNonce              string
	lastAuthorizationPath  string
	lastAuthorizationQuery url.Values
	lastDeviceForm         url.Values
	lastTokenForm          url.Values
	lastRevocation         url.Values
}

func newMockOIDCProvider(t *testing.T, includeUserInfo bool) *mockOIDCProvider {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, (&jose.SignerOptions{}).WithHeader("kid", "test-key").WithType("JWT"))
	require.NoError(t, err)

	jwkSet := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       &privateKey.PublicKey,
				KeyID:     "test-key",
				Algorithm: string(jose.RS256),
				Use:       "sig",
			},
		},
	}
	jwkJSON, err := json.Marshal(jwkSet)
	require.NoError(t, err)

	provider := &mockOIDCProvider{
		t:                t,
		signer:           signer,
		publicJWKSetJSON: jwkJSON,
		includeUserInfo:  includeUserInfo,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", provider.handleDiscovery)
	mux.HandleFunc("/authorize-default", provider.handleAuthorize)
	mux.HandleFunc("/authorize-override", provider.handleAuthorize)
	mux.HandleFunc("/device-code", provider.handleDeviceAuthorization)
	mux.HandleFunc("/token", provider.handleToken)
	mux.HandleFunc("/jwks", provider.handleJWKS)
	mux.HandleFunc("/userinfo", provider.handleUserInfo)
	mux.HandleFunc("/revoke", provider.handleRevoke)

	provider.server = httptest.NewServer(mux)
	return provider
}

func (p *mockOIDCProvider) Close() {
	p.server.Close()
}

func (p *mockOIDCProvider) IssuerURL() string {
	return p.server.URL
}

func (p *mockOIDCProvider) SetComputeAPIBaseURL(baseURL string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.computeAPIBaseURL = baseURL
}

func (p *mockOIDCProvider) LastAuthorizationPath() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastAuthorizationPath
}

func (p *mockOIDCProvider) LastAuthorizationQuery() url.Values {
	p.mu.Lock()
	defer p.mu.Unlock()
	return cloneValues(p.lastAuthorizationQuery)
}

func (p *mockOIDCProvider) LastTokenForm() url.Values {
	p.mu.Lock()
	defer p.mu.Unlock()
	return cloneValues(p.lastTokenForm)
}

func (p *mockOIDCProvider) LastDeviceAuthorizationForm() url.Values {
	p.mu.Lock()
	defer p.mu.Unlock()
	return cloneValues(p.lastDeviceForm)
}

func (p *mockOIDCProvider) LastRevocation() url.Values {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastRevocation
}

func (p *mockOIDCProvider) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	document := map[string]any{
		"issuer":                                p.server.URL,
		"authorization_endpoint":                p.server.URL + "/authorize-default",
		"device_authorization_endpoint":         p.server.URL + "/device-code",
		"token_endpoint":                        p.server.URL + "/token",
		"jwks_uri":                              p.server.URL + "/jwks",
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"revocation_endpoint":                   p.server.URL + "/revoke",
	}
	if p.includeUserInfo {
		document["userinfo_endpoint"] = p.server.URL + "/userinfo"
	}
	p.mu.Lock()
	computeAPIBaseURL := p.computeAPIBaseURL
	p.mu.Unlock()
	if computeAPIBaseURL != "" {
		document["boltz_compute_api_base_url"] = computeAPIBaseURL
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(document); err != nil {
		p.t.Errorf("encode discovery response: %v", err)
	}
	_ = r
}

func (p *mockOIDCProvider) handleDeviceAuthorization(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		p.t.Errorf("parse device authorization form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.mu.Lock()
	p.lastDeviceForm = cloneValues(r.PostForm)
	p.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"device_code":               "device-code-123",
		"user_code":                 "WDJB-MJHT",
		"verification_uri":          p.server.URL + "/device",
		"verification_uri_complete": p.server.URL + "/device?user_code=WDJB-MJHT",
		"expires_in":                600,
		"interval":                  1,
	}); err != nil {
		p.t.Errorf("encode device authorization response: %v", err)
	}
}

func (p *mockOIDCProvider) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.lastAuthorizationPath = r.URL.Path
	p.lastAuthorizationQuery = cloneValues(r.URL.Query())
	p.lastNonce = r.URL.Query().Get("nonce")
	p.mu.Unlock()

	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	callback, err := url.Parse(redirectURI)
	if err != nil {
		p.t.Errorf("parse redirect URI: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	query := callback.Query()
	query.Set("code", "login-code")
	query.Set("state", state)
	callback.RawQuery = query.Encode()

	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func (p *mockOIDCProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		p.t.Errorf("parse token form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.mu.Lock()
	p.lastTokenForm = cloneValues(r.PostForm)
	p.mu.Unlock()

	var response map[string]any
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		response = map[string]any{
			"access_token":  "login-access",
			"refresh_token": "login-refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid email",
			"id_token":      p.mustSignIDToken("client-123", p.currentNonce()),
		}
	case deviceCodeGrantType:
		require.Equal(p.t, "device-code-123", r.PostForm.Get("device_code"))
		response = map[string]any{
			"access_token":  "login-access",
			"refresh_token": "login-refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid email",
		}
	case "refresh_token":
		response = map[string]any{
			"access_token":  "refreshed-access",
			"refresh_token": "rotated-refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid email",
			"id_token":      p.mustSignIDToken("client-123", ""),
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"unsupported_grant_type"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		p.t.Errorf("encode token response: %v", err)
	}
}

func (p *mockOIDCProvider) handleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(p.publicJWKSetJSON)
	_ = r
}

func (p *mockOIDCProvider) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer login-access" {
		p.t.Errorf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"sub":"user-123","email":"userinfo@example.com","name":"Userinfo Name","preferred_username":"userinfo-user"}`))
}

func (p *mockOIDCProvider) handleRevoke(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		p.t.Errorf("parse revoke form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.mu.Lock()
	p.lastRevocation = r.PostForm
	p.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (p *mockOIDCProvider) currentNonce() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastNonce
}

func (p *mockOIDCProvider) mustSignIDToken(clientID, nonce string) string {
	claims := map[string]any{
		"iss":                p.server.URL,
		"sub":                "user-123",
		"aud":                clientID,
		"exp":                time.Now().Add(time.Hour).Unix(),
		"iat":                time.Now().Unix(),
		"email":              "user@example.com",
		"name":               "Test User",
		"preferred_username": "test-user",
	}
	if nonce != "" {
		claims["nonce"] = nonce
	}
	token, err := jwt.Signed(p.signer).Claims(claims).Serialize()
	if err != nil {
		panic(err)
	}
	return token
}

func visitURL(target string) {
	_, _ = http.Get(target)
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for key, value := range values {
		cloned[key] = append([]string(nil), value...)
	}
	return cloned
}
