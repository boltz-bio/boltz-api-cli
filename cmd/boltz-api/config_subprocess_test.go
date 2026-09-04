package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigHelpOmitsSetSubprocess(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	result := runCLI(t, binary, env, "config", "--help")
	require.Equal(t, 0, result.ExitCode, result.Stderr)
	require.Contains(t, result.Stdout, "show")
	require.Contains(t, result.Stdout, "reset")
	require.NotContains(t, result.Stdout, "\n   set ")
	require.NotContains(t, result.Stdout, "Persist local CLI configuration")

	rootHelp := runCLI(t, binary, env, "--help")
	require.Equal(t, 0, rootHelp.ExitCode, rootHelp.Stderr)
	require.Contains(t, rootHelp.Stdout, "OIDC issuer URL for OAuth login, tenant API discovery, and bearer-token refresh")
}
