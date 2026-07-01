// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/boltz-bio/boltz-api-cli/internal/autocomplete"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
)

var (
	Command            *cli.Command
	CommandErrorBuffer bytes.Buffer
)

func init() {
	Command = &cli.Command{
		Name:      "boltz-api",
		Usage:     "CLI for the boltz API",
		Suggest:   true,
		Version:   Version,
		ErrWriter: &CommandErrorBuffer,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug logging",
			},
			&cli.StringFlag{
				Name:        "base-url",
				DefaultText: "url",
				Usage:       "Override the base URL for API requests",
				Validator: func(baseURL string) error {
					return ValidateBaseURL(baseURL, "--base-url")
				},
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "The format for displaying response data (one of: " + strings.Join(OutputFormats, ", ") + ")",
				Value: "auto",
				Validator: func(format string) error {
					if !slices.Contains(OutputFormats, strings.ToLower(format)) {
						return fmt.Errorf("format must be one of: %s", strings.Join(OutputFormats, ", "))
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:  "format-error",
				Usage: "The format for displaying error data (one of: " + strings.Join(OutputFormats, ", ") + ")",
				Value: "auto",
				Validator: func(format string) error {
					if !slices.Contains(OutputFormats, strings.ToLower(format)) {
						return fmt.Errorf("format must be one of: %s", strings.Join(OutputFormats, ", "))
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:  "transform",
				Usage: "The GJSON transformation for data output.",
			},
			&cli.StringFlag{
				Name:  "transform-error",
				Usage: "The GJSON transformation for errors.",
			},
			&cli.BoolFlag{
				Name:    "raw-output",
				Aliases: []string{"r"},
				Usage:   "If the result is a string, print it without JSON quotes. This can be useful for making output transforms talk to non-JSON-based systems.",
			},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_API_KEY"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "predictions:structure-and-binding",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&predictionsStructureAndBindingRetrieve,
					&predictionsStructureAndBindingList,
					&predictionsStructureAndBindingDeleteData,
					&predictionsStructureAndBindingEstimateCost,
					&predictionsStructureAndBindingStart,
				},
			},
			{
				Name:     "predictions:adme",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&predictionsAdmeRetrieve,
					&predictionsAdmeList,
					&predictionsAdmeDeleteData,
					&predictionsAdmeEstimateCost,
					&predictionsAdmeStart,
				},
			},
			{
				Name:     "small-molecule:design",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&smallMoleculeDesignRetrieve,
					&smallMoleculeDesignList,
					&smallMoleculeDesignDeleteData,
					&smallMoleculeDesignEstimateCost,
					&smallMoleculeDesignListResults,
					&smallMoleculeDesignResume,
					&smallMoleculeDesignStart,
					&smallMoleculeDesignStop,
				},
			},
			{
				Name:     "small-molecule:library-screen",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&smallMoleculeLibraryScreenRetrieve,
					&smallMoleculeLibraryScreenList,
					&smallMoleculeLibraryScreenDeleteData,
					&smallMoleculeLibraryScreenEstimateCost,
					&smallMoleculeLibraryScreenListResults,
					&smallMoleculeLibraryScreenResume,
					&smallMoleculeLibraryScreenStart,
					&smallMoleculeLibraryScreenStop,
				},
			},
			{
				Name:     "protein:design",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&proteinDesignRetrieve,
					&proteinDesignList,
					&proteinDesignDeleteData,
					&proteinDesignEstimateCost,
					&proteinDesignListResults,
					&proteinDesignResume,
					&proteinDesignStart,
					&proteinDesignStop,
				},
			},
			{
				Name:     "protein:library-screen",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&proteinLibraryScreenRetrieve,
					&proteinLibraryScreenList,
					&proteinLibraryScreenDeleteData,
					&proteinLibraryScreenEstimateCost,
					&proteinLibraryScreenListResults,
					&proteinLibraryScreenResume,
					&proteinLibraryScreenStart,
					&proteinLibraryScreenStop,
				},
			},
			{
				Name:     "admin:workspaces",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&adminWorkspacesCreate,
					&adminWorkspacesRetrieve,
					&adminWorkspacesUpdate,
					&adminWorkspacesList,
					&adminWorkspacesArchive,
					&adminWorkspacesRetrieveSpendingLimit,
					&adminWorkspacesSetSpendingLimit,
				},
			},
			{
				Name:     "admin:api-keys",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&adminAPIKeysCreate,
					&adminAPIKeysList,
					&adminAPIKeysRevoke,
				},
			},
			{
				Name:     "admin:usage",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&adminUsageList,
				},
			},
			{
				Name:     "cli",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&cliVersion,
				},
			},
			{
				Name:     "auth",
				Category: "API RESOURCE",
				Suggest:  true,
				Commands: []*cli.Command{
					&authMe,
				},
			},
			{
				Name:            "@manpages",
				Usage:           "Generate documentation for 'man'",
				UsageText:       "boltz-api @manpages [-o boltz-api.1] [--gzip]",
				Hidden:          true,
				Action:          generateManpages,
				HideHelpCommand: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "write manpages to the given folder",
						Value:   "man",
					},
					&cli.BoolFlag{
						Name:    "gzip",
						Aliases: []string{"z"},
						Usage:   "output gzipped manpage files to .gz",
						Value:   true,
					},
					&cli.BoolFlag{
						Name:    "text",
						Aliases: []string{"z"},
						Usage:   "output uncompressed text files",
						Value:   false,
					},
				},
			},
			{
				Name:            "__complete",
				Hidden:          true,
				HideHelpCommand: true,
				Action:          autocomplete.ExecuteShellCompletion,
			},
			{
				Name:            "@completion",
				Hidden:          true,
				HideHelpCommand: true,
				Action:          autocomplete.OutputCompletionScript,
			},
		},
		HideHelpCommand: true,
	}

	// Temporary Stainless seam: register custom commands, flags, and tree-level
	// post-processing in one place.
	ApplyCustomizations(Command)
}

func generateManpages(ctx context.Context, c *cli.Command) error {
	manpage, err := docs.ToManWithSection(Command, 1)
	if err != nil {
		return err
	}
	dir := c.String("output")
	err = os.MkdirAll(filepath.Join(dir, "man1"), 0755)
	if err != nil {
		// handle error
	}
	if c.Bool("text") {
		file, err := os.Create(filepath.Join(dir, "man1", "boltz-api.1"))
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := file.WriteString(manpage); err != nil {
			return err
		}
	}
	if c.Bool("gzip") {
		file, err := os.Create(filepath.Join(dir, "man1", "boltz-api.1.gz"))
		if err != nil {
			return err
		}
		defer file.Close()
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		_, err = gzWriter.Write([]byte(manpage))
		if err != nil {
			return err
		}
	}
	fmt.Printf("Wrote manpages to %s\n", dir)
	return nil
}
