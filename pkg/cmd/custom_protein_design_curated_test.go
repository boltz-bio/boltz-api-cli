// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestApplyCustomizationsAddsProteinDesignListCuratedSpecifications(t *testing.T) {
	root := newProteinDesignCuratedTestRoot(io.Discard)

	ApplyCustomizations(root)
	ApplyCustomizations(root)

	parent := mustFindCommand(t, root, "protein:design")
	command := mustFindCommand(t, parent, "list-curated-specifications")
	require.NotNil(t, findFlag(command, "type"))
	require.Len(t, matchingCommands(parent.Commands, "list-curated-specifications"), 1)
}

func TestProteinDesignListCuratedSpecificationsSendsType(t *testing.T) {
	setAuthCommandUserDirs(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/compute/v1/protein/design/curated-specifications", r.URL.Path)
		require.Equal(t, "nanobody", r.URL.Query().Get("type"))
		require.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	root := newProteinDesignCuratedTestRoot(io.Discard)
	require.NoError(t, root.Run(context.Background(), []string{
		"boltz-api",
		"--base-url", server.URL,
		"--api-key", "test-api-key",
		"--format", "json",
		"protein:design", "list-curated-specifications",
		"--type", "nanobody",
	}))
}

func TestProteinDesignListCuratedSpecificationsRejectsUnknownType(t *testing.T) {
	root := newProteinDesignCuratedTestRoot(io.Discard)
	err := root.Run(context.Background(), []string{
		"boltz-api",
		"protein:design", "list-curated-specifications",
		"--type", "unknown",
	})

	require.ErrorContains(t, err, "type must be one of: nanobody, antibody")
}

func newProteinDesignCuratedTestRoot(writer io.Writer) *cli.Command {
	root := &cli.Command{
		Name:      "boltz-api",
		Writer:    writer,
		ErrWriter: writer,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url"},
			&cli.StringFlag{Name: "format", Value: "auto"},
			&cli.BoolFlag{Name: "raw-output"},
			&cli.StringFlag{Name: "transform"},
			&requestflag.Flag[string]{Name: "api-key"},
		},
		Commands: []*cli.Command{
			{
				Name: "protein:design",
				Commands: []*cli.Command{
					{Name: "list"},
				},
			},
		},
	}
	ApplyCustomizations(root)
	return root
}

func matchingCommands(commands []*cli.Command, name string) []*cli.Command {
	matches := make([]*cli.Command, 0, 1)
	for _, command := range commands {
		if command != nil && command.Name == name {
			matches = append(matches, command)
		}
	}
	return matches
}
