// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/autherror"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/urfave/cli/v3"
)

var configCommand = &cli.Command{
	Name:            "config",
	Usage:           "Inspect and reset local CLI configuration",
	Suggest:         true,
	HideHelpCommand: true,
	Commands: []*cli.Command{
		{
			Name:            "show",
			Usage:           "Show the local non-secret configuration file",
			Action:          handleConfigShow,
			HideHelpCommand: true,
		},
		{
			Name:            "reset",
			Usage:           "Remove the local non-secret configuration file",
			Action:          handleConfigReset,
			HideHelpCommand: true,
		},
	},
}

type configShowResponse struct {
	Path    string              `json:"path"`
	Present bool                `json:"present"`
	Config  *configFileResponse `json:"config,omitempty"`
}

type configFileResponse struct {
	Version          int      `json:"version,omitempty"`
	BaseURL          string   `json:"base_url,omitempty"`
	IssuerURL        string   `json:"issuer_url,omitempty"`
	ClientID         string   `json:"client_id,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
	Audience         string   `json:"audience,omitempty"`
	AuthorizationURL string   `json:"authorization_url,omitempty"`
	TokenURL         string   `json:"token_url,omitempty"`
	UserInfoURL      string   `json:"userinfo_url,omitempty"`
	RevocationURL    string   `json:"revocation_url,omitempty"`
	SelectedOrg      string   `json:"selected_org,omitempty"`
}

type configResetResponse struct {
	Path    string `json:"path"`
	Removed bool   `json:"removed"`
}

func handleConfigShow(ctx context.Context, cmd *cli.Command) error {
	_ = ctx

	path, err := authstore.ConfigFilePath()
	if err != nil {
		return authmodeConfigError(err)
	}
	config, err := authconfig.Load()
	if err != nil {
		return authmodeConfigError(err)
	}
	present, err := fileExists(path)
	if err != nil {
		return authmodeConfigError(err)
	}

	response := configShowResponse{
		Path:    path,
		Present: present,
	}
	if present {
		response.Config = &configFileResponse{
			Version:          config.Version,
			BaseURL:          config.BaseURL,
			IssuerURL:        config.IssuerURL,
			ClientID:         config.ClientID,
			Scopes:           append([]string(nil), config.Scopes...),
			Audience:         config.Audience,
			AuthorizationURL: config.AuthorizationURL,
			TokenURL:         config.TokenURL,
			UserInfoURL:      config.UserInfoURL,
			RevocationURL:    config.RevocationURL,
			SelectedOrg:      config.SelectedOrg,
		}
	}
	return showJSONValue(cmd, response, "config show")
}

func handleConfigReset(ctx context.Context, cmd *cli.Command) error {
	_ = ctx

	path, err := authstore.ConfigFilePath()
	if err != nil {
		return authmodeConfigError(err)
	}
	removed := true
	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		removed = false
	} else if err != nil {
		return authmodeConfigError(err)
	}

	return showJSONValue(cmd, configResetResponse{Path: path, Removed: removed}, "config reset")
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func authmodeConfigError(err error) error {
	return autherror.New("config_error", "Failed to read local CLI configuration", err.Error())
}
