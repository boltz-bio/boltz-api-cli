// Custom CLI extension code. Not generated.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"

	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestApplyCustomizationsPreservesNativeIDFlags(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	for _, tc := range []struct {
		path []string
	}{
		{[]string{"small-molecule:design", "retrieve"}},
		{[]string{"small-molecule:library-screen", "list-results"}},
		{[]string{"protein:design", "stop"}},
		{[]string{"protein:library-screen", "delete-data"}},
	} {
		cmd := mustFindCommand(t, Command, tc.path...)
		mustFindFlag(t, cmd, "id")
		require.Nil(t, findFlag(cmd, "run-id"))
		require.Nil(t, findFlag(cmd, "screen-id"))
	}
}

func TestApplyCustomizationsIsIdempotent(t *testing.T) {
	t.Parallel()

	root := &cli.Command{
		Name: "boltz-api",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "transform"},
		},
	}

	ApplyCustomizations(root)
	ApplyCustomizations(root)

	require.Len(t, root.Commands, 4)
	require.Len(t, root.Flags, 11)
	require.Equal(t, transformUsage, usageForFlag(t, mustFindFlag(t, root, "transform")))
	require.Nil(t, findFlag(root, "no-browser"))
	mustFindCommand(t, root, "auth")
	mustFindCommand(t, root, "config")
	mustFindCommand(t, root, "download-results")
	mustFindCommand(t, root, "download-status")
}

func TestApplyCustomizationsMergesGeneratedAuthCommand(t *testing.T) {
	root := &cli.Command{
		Name: "boltz-api",
		Commands: []*cli.Command{
			{
				Name: "auth",
				Commands: []*cli.Command{
					{
						Name: "me",
					},
				},
			},
		},
	}

	ApplyCustomizations(root)
	ApplyCustomizations(root)

	require.Len(t, root.Commands, 4)
	auth := mustFindCommand(t, root, "auth")
	mustFindCommand(t, auth, "me")
	mustFindCommand(t, auth, "login")
	mustFindCommand(t, auth, "status")
	mustFindCommand(t, auth, "validate")
	mustFindCommand(t, auth, "orgs")
	mustFindCommand(t, auth, "switch-org")
	require.Len(t, auth.Commands, len(authCommand.Commands)+1)
}

func TestApplyCustomizationsAnnotatesRepeatableArrayFlags(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	moleculeFlag := mustFindFlag(t, mustFindCommand(t, Command, "small-molecule:library-screen", "start"), "molecule")
	require.Contains(t, usageForFlag(t, moleculeFlag), "Repeat --molecule to add entries.")
	require.Contains(t, usageForFlag(t, moleculeFlag), "use the body field molecules")

	proteinFlag := mustFindFlag(t, mustFindCommand(t, Command, "protein:library-screen", "start"), "protein")
	require.Contains(t, usageForFlag(t, proteinFlag), "Repeat --protein to add entries.")
	require.Contains(t, usageForFlag(t, proteinFlag), "use the body field proteins")
}

func TestApplyCustomizationsAddsMergedInputFlags(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	for _, tc := range []struct {
		path       []string
		innerFlags []string
	}{
		{
			path:       []string{"small-molecule:design", "estimate-cost"},
			innerFlags: []string{"input.num-molecules", "input.target", "input.chemical-space", "input.molecule-filters"},
		},
		{
			path:       []string{"small-molecule:design", "start"},
			innerFlags: []string{"input.num-molecules", "input.target", "input.chemical-space", "input.molecule-filters"},
		},
		{
			path:       []string{"small-molecule:library-screen", "estimate-cost"},
			innerFlags: []string{"input.molecules", "input.target", "input.molecule-filters"},
		},
		{
			path:       []string{"small-molecule:library-screen", "start"},
			innerFlags: []string{"input.molecules", "input.target", "input.molecule-filters"},
		},
		{
			path:       []string{"protein:design", "estimate-cost"},
			innerFlags: []string{"input.binder-specification", "input.num-proteins", "input.target"},
		},
		{
			path:       []string{"protein:design", "start"},
			innerFlags: []string{"input.binder-specification", "input.num-proteins", "input.target"},
		},
		{
			path:       []string{"protein:library-screen", "estimate-cost"},
			innerFlags: []string{"input.proteins", "input.target"},
		},
		{
			path:       []string{"protein:library-screen", "start"},
			innerFlags: []string{"input.proteins", "input.target"},
		},
	} {
		cmd := mustFindCommand(t, Command, tc.path...)
		inputFlag := mustFindFlag(t, cmd, "input")
		require.Contains(t, cmd.Usage, "Prefer `--input` for full payloads")
		require.Contains(t, usageForFlag(t, inputFlag), "@json://")
		require.Contains(t, usageForFlag(t, inputFlag), "--idempotency-key")
		require.Contains(t, usageForFlag(t, inputFlag), "--workspace-id")

		for _, name := range tc.innerFlags {
			mustFindFlag(t, cmd, name)
		}

		switch tc.path[0] {
		case "small-molecule:design", "small-molecule:library-screen", "protein:design", "protein:library-screen":
			targetFlag := findFlag(cmd, "target")
			if targetFlag != nil {
				require.Contains(t, usageForFlag(t, targetFlag), "prefer `--input` for the full payload")
				require.Contains(t, usageForFlag(t, targetFlag), "overrides the matching field from `--input`")
			}
		}

		workspaceFlag := findFlag(cmd, "workspace-id")
		if workspaceFlag != nil {
			require.Contains(t, usageForFlag(t, workspaceFlag), "Keep this as a top-level flag")
			require.Contains(t, usageForFlag(t, workspaceFlag), "top-level flag wins")
		}

		require.NoError(t, requestflag.CheckInnerFlags(*cmd))
	}
}

func TestAddMergedInputFlagDerivesInnerFlagsFromTopLevelBodyFlags(t *testing.T) {
	t.Parallel()

	cmd := &cli.Command{
		Name: "start",
		Flags: []cli.Flag{
			&requestflag.Flag[string]{
				Name:     "priority",
				Usage:    "Synthetic future top-level field.",
				BodyPath: "priority",
			},
			&requestflag.Flag[string]{
				Name:     "workspace-id",
				BodyPath: "workspace_id",
			},
			&requestflag.Flag[string]{
				Name:     "idempotency-key",
				BodyPath: "idempotency_key",
			},
		},
	}

	addMergedInputFlag(cmd, mergedInputUsage)

	mustFindFlag(t, cmd, "input")
	mustFindFlag(t, cmd, "input.priority")
	require.Nil(t, findFlag(cmd, "input.workspace-id"))
	require.Nil(t, findFlag(cmd, "input.idempotency-key"))
	require.NoError(t, requestflag.CheckInnerFlags(*cmd))
}

func TestCustomizeProteinDesignModeFlagsSupportsLegacyAndDiscriminatedRequests(t *testing.T) {
	t.Parallel()

	for _, commandName := range []string{"estimate-cost", "start"} {
		commandName := commandName
		t.Run(commandName, func(t *testing.T) {
			t.Parallel()

			modeFlag := &requestflag.Flag[string]{
				Name:     "type",
				Usage:    `Allowed values: "binder".`,
				Default:  "binder",
				Const:    true,
				BodyPath: "type",
			}
			requestCommand := &cli.Command{
				Name:  commandName,
				Flags: []cli.Flag{modeFlag},
			}
			root := &cli.Command{
				Name: "boltz-api",
				Commands: []*cli.Command{
					{
						Name:     "protein:design",
						Commands: []*cli.Command{requestCommand},
					},
				},
			}

			customizeProteinDesignModeFlags(root)

			require.Empty(t, modeFlag.Default)
			require.Empty(t, modeFlag.DefaultText)
			require.False(t, modeFlag.Const)
			require.Equal(t, proteinDesignModeUsage, modeFlag.Usage)
			require.NoError(t, modeFlag.Validator(""))
			require.NoError(t, modeFlag.Validator("binder"))
			require.NoError(t, modeFlag.Validator("generic"))
			require.EqualError(t, modeFlag.Validator("invalid"), "type must be one of: binder, generic")

			require.NoError(t, modeFlag.PreParse())
			require.False(t, modeFlag.IsSet(), "legacy requests must omit the discriminator")
		})
	}
}

func TestGeneratedProteinDesignCommandsExposeBothRequestModes(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	for _, commandName := range []string{"estimate-cost", "start"} {
		requestCommand := mustFindCommand(t, Command, "protein:design", commandName)
		modeFlag, ok := mustFindFlag(t, requestCommand, "type").(*requestflag.Flag[string])
		require.True(t, ok)
		require.Empty(t, modeFlag.Default)
		require.False(t, modeFlag.Const)
		require.Contains(t, modeFlag.Usage, proteinDesignModeUsage)
		mustFindFlag(t, requestCommand, "binder")
		mustFindFlag(t, requestCommand, "entity")
		mustFindFlag(t, requestCommand, "input.type")
	}
}

func TestProteinDesignModeFlagDoesNotOverwriteMergedInput(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		input    string
		wantType any
	}{
		{
			name:     "legacy request omits discriminator",
			input:    `{"binder_specification":{"type":"no_template"},"target":{"sequence":"ACDE"}}`,
			wantType: nil,
		},
		{
			name:     "generic request keeps input discriminator",
			input:    `{"type":"generic","entities":[]}`,
			wantType: "generic",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			modeFlag := &requestflag.Flag[string]{
				Name:     "type",
				Default:  "binder",
				Const:    true,
				BodyPath: "type",
			}
			requestCommand := &cli.Command{Name: "start", Flags: []cli.Flag{modeFlag}}
			root := &cli.Command{
				Commands: []*cli.Command{{
					Name:     "protein:design",
					Commands: []*cli.Command{requestCommand},
				}},
			}

			customizeProteinDesignModeFlags(root)
			addMergedInputFlag(requestCommand, mergedInputUsage)

			inputFlag, ok := findFlag(requestCommand, "input").(*requestflag.Flag[map[string]any])
			require.True(t, ok)
			require.NoError(t, inputFlag.PreParse())
			require.NoError(t, modeFlag.PreParse())
			require.NoError(t, inputFlag.Set("input", tc.input))

			body, ok := requestflag.ExtractRequestContents(requestCommand).Body.(map[string]any)
			require.True(t, ok)
			if tc.wantType == nil {
				require.NotContains(t, body, "type")
			} else {
				require.Equal(t, tc.wantType, body["type"])
			}
		})
	}
}

func TestTransformUsageMentionsPerItemListBehavior(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	transformFlag := mustFindFlag(t, Command, "transform")
	require.Contains(t, usageForFlag(t, transformFlag), "runs on each emitted item")
	require.Contains(t, usageForFlag(t, transformFlag), "--format raw")
	require.Contains(t, usageForFlag(t, transformFlag), "jq")
}

func TestManpageIncludesRepeatableArrayFlagGuidance(t *testing.T) {
	t.Parallel()

	ApplyCustomizations(Command)

	manpage, err := docs.ToManWithSection(Command, 1)
	require.NoError(t, err)
	require.Contains(t, manpage, "Repeat --molecule to add entries")
	require.Contains(t, manpage, "use the body field molecules")
	require.Contains(t, manpage, "Repeat --protein to add entries")
	require.Contains(t, manpage, "use the body field proteins")
	require.Contains(t, manpage, "top-level flags remain available as overrides")
	require.Contains(t, manpage, "@json://...")
	require.Contains(t, manpage, "Keep this as a top-level flag even when using")
}

func TestRepeatedSingularFlagsPopulatePluralBodyField(t *testing.T) {
	t.Parallel()

	moleculeFlag := &requestflag.Flag[[]map[string]any]{
		Name:     "molecule",
		BodyPath: "molecules",
	}
	require.NoError(t, moleculeFlag.PreParse())
	require.NoError(t, moleculeFlag.Set("molecule", "{smiles: CCO}"))
	require.NoError(t, moleculeFlag.Set("molecule", "{smiles: CCN}"))

	cmd := &cli.Command{
		Name:  "start",
		Flags: []cli.Flag{moleculeFlag},
	}

	body, ok := requestflag.ExtractRequestContents(cmd).Body.(map[string]any)
	require.True(t, ok)

	molecules, ok := body["molecules"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, molecules, 2)
	require.Equal(t, "CCO", molecules[0]["smiles"])
	require.Equal(t, "CCN", molecules[1]["smiles"])
}

func mustFindCommand(t *testing.T, root *cli.Command, path ...string) *cli.Command {
	t.Helper()

	cmd := root
	for _, part := range path {
		var next *cli.Command
		for _, child := range cmd.Commands {
			if child != nil && child.Name == part {
				next = child
				break
			}
		}
		require.NotNilf(t, next, "command %q not found under %q", part, cmd.Name)
		cmd = next
	}
	return cmd
}

func mustFindFlag(t *testing.T, cmd *cli.Command, name string) cli.Flag {
	t.Helper()

	flag := findFlag(cmd, name)
	if flag != nil {
		return flag
	}
	t.Fatalf("flag %q not found on command %q", name, cmd.Name)
	return nil
}

func findFlag(cmd *cli.Command, name string) cli.Flag {
	for _, flag := range cmd.Flags {
		if len(flag.Names()) > 0 && flag.Names()[0] == name {
			return flag
		}
	}
	return nil
}

func usageForFlag(t *testing.T, flag cli.Flag) string {
	t.Helper()

	docFlag, ok := flag.(interface{ GetUsage() string })
	require.True(t, ok)
	return docFlag.GetUsage()
}
