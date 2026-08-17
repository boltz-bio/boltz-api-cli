package authconfig

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

const (
	ConfigVersion           = 1
	EnvAuthIssuerURL        = "BOLTZ_API_AUTH_ISSUER_URL"
	EnvAuthClientID         = "BOLTZ_API_AUTH_CLIENT_ID"
	EnvAuthScope            = "BOLTZ_API_AUTH_SCOPE"
	EnvAuthAudience         = "BOLTZ_API_AUTH_AUDIENCE"
	EnvAuthAuthorizationURL = "BOLTZ_API_AUTH_AUTHORIZATION_URL"
	EnvAuthTokenURL         = "BOLTZ_API_AUTH_TOKEN_URL"
	EnvAuthUserInfoURL      = "BOLTZ_API_AUTH_USERINFO_URL"
	EnvAuthRevocationURL    = "BOLTZ_API_AUTH_REVOCATION_URL"
	EnvOrg                  = "BOLTZ_API_ORG"
	EnvListenPort           = "BOLTZ_API_LISTEN_PORT"
	EnvBaseURL              = "BOLTZ_BASE_URL"
)

var DefaultScopes = []string{"openid", "offline_access", "profile", "email"}

const (
	DefaultIssuerURL  = "https://lab.boltz.bio"
	DefaultClientID   = "boltz-cli"
	DefaultAudience   = "boltz-compute-api"
	DefaultListenPort = 8421
	DefaultScope      = "compute:run"
)

type Source string

const (
	SourceUnset        Source = "unset"
	SourceRuntime      Source = "runtime"
	SourceConfig       Source = "config"
	SourceDefault      Source = "default"
	SourceSessionCache Source = "session_cache"
	SourceKeyring      Source = "keyring"
	SourceFile         Source = "file"
)

type FileConfig struct {
	Version          int      `yaml:"version,omitempty"`
	BaseURL          string   `yaml:"base_url,omitempty"`
	IssuerURL        string   `yaml:"issuer_url,omitempty"`
	ClientID         string   `yaml:"client_id,omitempty"`
	Scopes           []string `yaml:"scopes,omitempty"`
	Audience         string   `yaml:"audience,omitempty"`
	AuthorizationURL string   `yaml:"authorization_url,omitempty"`
	TokenURL         string   `yaml:"token_url,omitempty"`
	UserInfoURL      string   `yaml:"userinfo_url,omitempty"`
	RevocationURL    string   `yaml:"revocation_url,omitempty"`
	SelectedOrg      string   `yaml:"selected_org,omitempty"`
}

type Resolved struct {
	APIKey           string
	BaseURL          string
	IssuerURL        string
	ClientID         string
	Scopes           []string
	Audience         string
	AuthorizationURL string
	TokenURL         string
	UserInfoURL      string
	RevocationURL    string
	SelectedOrg      string
	ListenPort       int
	Sources          Sources
}

type Sources struct {
	APIKey           Source
	BaseURL          Source
	IssuerURL        Source
	ClientID         Source
	Scopes           Source
	Audience         Source
	AuthorizationURL Source
	TokenURL         Source
	UserInfoURL      Source
	RevocationURL    Source
	SelectedOrg      Source
	ListenPort       Source
}

func Load() (FileConfig, error) {
	path, err := authstore.ConfigFilePath()
	if err != nil {
		return FileConfig{}, err
	}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return FileConfig{}, nil
	}
	if err != nil {
		return FileConfig{}, err
	}

	var config FileConfig
	if err := yaml.Unmarshal(body, &config); err != nil {
		return FileConfig{}, err
	}
	config.Scopes = normalizeScopes(config.Scopes)
	return config, nil
}

func Resolve(cmd *cli.Command) (Resolved, error) {
	root := cmd.Root()
	if root == nil {
		root = cmd
	}

	config, err := Load()
	if err != nil {
		return Resolved{}, err
	}

	apiKey, apiKeySource := resolveAPIKey(root)
	baseURL, baseURLSource, err := resolveBaseURL(root, config.BaseURL)
	if err != nil {
		return Resolved{}, err
	}
	issuerURL, issuerSource := resolveString(root, config.IssuerURL, "auth-issuer-url")
	if issuerURL == "" {
		issuerURL = DefaultIssuerURL
		issuerSource = SourceDefault
	}
	clientID, clientIDSource := resolveString(root, config.ClientID, "auth-client-id")
	if clientID == "" {
		clientID = DefaultClientID
		clientIDSource = SourceDefault
	}
	scopes, scopesSource := resolveScopes(root, config.Scopes)
	audience, audienceSource := resolveString(root, config.Audience, "auth-audience")
	if audience == "" {
		audience = DefaultAudience
		audienceSource = SourceDefault
	}
	authorizationURL, authorizationSource := resolveString(root, config.AuthorizationURL, "auth-authorization-url")
	tokenURL, tokenURLSource := resolveString(root, config.TokenURL, "auth-token-url")
	userInfoURL, userInfoURLSource := resolveString(root, config.UserInfoURL, "auth-userinfo-url")
	revocationURL, revocationSource := resolveString(root, config.RevocationURL, "auth-revocation-url")
	selectedOrg, selectedOrgSource := resolveString(root, config.SelectedOrg, "org")
	listenPort, listenPortSource := resolveInt(root, DefaultListenPort, "listen-port")

	return Resolved{
		APIKey:           apiKey,
		BaseURL:          baseURL,
		IssuerURL:        issuerURL,
		ClientID:         clientID,
		Scopes:           scopes,
		Audience:         audience,
		AuthorizationURL: authorizationURL,
		TokenURL:         tokenURL,
		UserInfoURL:      userInfoURL,
		RevocationURL:    revocationURL,
		SelectedOrg:      selectedOrg,
		ListenPort:       listenPort,
		Sources: Sources{
			APIKey:           apiKeySource,
			BaseURL:          baseURLSource,
			IssuerURL:        issuerSource,
			ClientID:         clientIDSource,
			Scopes:           scopesSource,
			Audience:         audienceSource,
			AuthorizationURL: authorizationSource,
			TokenURL:         tokenURLSource,
			UserInfoURL:      userInfoURLSource,
			RevocationURL:    revocationSource,
			SelectedOrg:      selectedOrgSource,
			ListenPort:       listenPortSource,
		},
	}, nil
}

func SaveProfile(resolved Resolved) error {
	return update(func(config *FileConfig) {
		// BaseURL is intentionally left untouched: --base-url and BOLTZ_BASE_URL are
		// invocation-scoped, while only `config set` should change the stored value.
		config.IssuerURL = strings.TrimSpace(resolved.IssuerURL)
		config.ClientID = strings.TrimSpace(resolved.ClientID)
		config.Scopes = normalizeScopes(resolved.Scopes)
		config.Audience = strings.TrimSpace(resolved.Audience)
		config.AuthorizationURL = strings.TrimSpace(resolved.AuthorizationURL)
		config.TokenURL = strings.TrimSpace(resolved.TokenURL)
		config.UserInfoURL = strings.TrimSpace(resolved.UserInfoURL)
		config.RevocationURL = strings.TrimSpace(resolved.RevocationURL)
		config.SelectedOrg = strings.TrimSpace(resolved.SelectedOrg)
	})
}

// SaveBaseURL updates the persistent API endpoint without changing auth fields.
func SaveBaseURL(baseURL string) error {
	return update(func(config *FileConfig) {
		config.BaseURL = strings.TrimSpace(baseURL)
	})
}

func SaveSelectedOrg(org string) error {
	return update(func(config *FileConfig) {
		config.SelectedOrg = strings.TrimSpace(org)
	})
}

// update applies mutate to the stored config, preserving every field the
// caller does not own.
func update(mutate func(*FileConfig)) error {
	config, err := Load()
	if err != nil {
		return err
	}
	mutate(&config)
	return save(config)
}

func save(config FileConfig) error {
	config.Version = ConfigVersion
	config.Scopes = normalizeScopes(config.Scopes)

	path, err := authstore.ConfigFilePath()
	if err != nil {
		return err
	}
	body, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return authstore.WriteFileAtomically(path, body, 0o600)
}

func resolveAPIKey(root *cli.Command) (string, Source) {
	value := strings.TrimSpace(root.String("api-key"))
	if root.IsSet("api-key") && value != "" {
		return value, SourceRuntime
	}
	if value != "" {
		return value, SourceRuntime
	}
	return "", SourceUnset
}

// ResolveBaseURL resolves only the effective API base URL (flag, env, stored
// config) for callers that don't need the full auth profile. A nil cmd skips
// the flag source.
func ResolveBaseURL(cmd *cli.Command) (string, Source, error) {
	var root *cli.Command
	if cmd != nil {
		root = cmd.Root()
		if root == nil {
			root = cmd
		}
	}
	config, err := Load()
	if err != nil {
		return "", SourceUnset, err
	}
	return resolveBaseURL(root, config.BaseURL)
}

func resolveBaseURL(root *cli.Command, fallback string) (string, Source, error) {
	if root != nil && root.IsSet("base-url") {
		value := strings.TrimSpace(root.String("base-url"))
		if value != "" {
			return value, SourceRuntime, ValidateBaseURL(value, "--base-url")
		}
	}
	if value := strings.TrimSpace(os.Getenv(EnvBaseURL)); value != "" {
		return value, SourceRuntime, ValidateBaseURL(value, EnvBaseURL)
	}
	if value := strings.TrimSpace(fallback); value != "" {
		return value, SourceConfig, ValidateBaseURL(value, "config base_url")
	}
	return "", SourceDefault, nil
}

// ValidateBaseURL preserves the CLI's existing base-URL validation behavior.
func ValidateBaseURL(value, source string) error {
	if value != "" && !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return fmt.Errorf("%s %q is missing a scheme (expected http:// or https://)", source, value)
	}
	return nil
}

func resolveString(root *cli.Command, fallback string, name string) (string, Source) {
	if root.IsSet(name) {
		return strings.TrimSpace(root.String(name)), SourceRuntime
	}
	value := strings.TrimSpace(fallback)
	if value != "" {
		return value, SourceConfig
	}
	return "", SourceUnset
}

func resolveInt(root *cli.Command, fallback int, name string) (int, Source) {
	if root.IsSet(name) {
		return root.Int(name), SourceRuntime
	}
	return fallback, SourceDefault
}

func resolveScopes(root *cli.Command, fallback []string) ([]string, Source) {
	if root.IsSet("auth-scope") {
		scopes := normalizeScopes(root.StringSlice("auth-scope"))
		if len(scopes) > 0 {
			return scopes, SourceRuntime
		}
	}

	scopes := normalizeScopes(fallback)
	if len(scopes) > 0 {
		return scopes, SourceConfig
	}
	return append(slices.Clone(DefaultScopes), DefaultScope), SourceDefault
}

func normalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}

	result := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, raw := range scopes {
		for _, part := range strings.Split(raw, ",") {
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
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
