// Custom CLI extension code. Not generated.
package cmd

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/urfave/cli/v3"
)

const (
	customizationsAppliedKey = "command_customizations_applied"
	transformUsage           = "The GJSON transformation for data output. For paginated or streamed list commands, this runs on each emitted item except in --format raw, where it runs on the full response page; use jq for whole-list reshaping."
	mergedInputCommandNote   = "Prefer `--input` for full payloads; top-level flags remain available as overrides."
	mergedInputFlagNote      = "Prefer `--input` for complete structured payloads or `@json://...` files. Top-level body flags can still override individual fields. Keep `--idempotency-key` and `--workspace-id` top-level; if also present in `--input`, the top-level flag wins."
	mergedFieldOverrideNote  = "You can set this directly, or prefer `--input` for the full payload. If both are set, this flag overrides the matching field from `--input`."
	mergedReservedFieldNote  = "Keep this as a top-level flag even when using `--input`. If also present in `--input`, this top-level flag wins."
)

func ApplyCustomizations(app *cli.Command) {
	if app == nil {
		return
	}
	if app.Metadata == nil {
		app.Metadata = map[string]any{}
	}
	if applied, _ := app.Metadata[customizationsAppliedKey].(bool); applied {
		return
	}
	app.Metadata[customizationsAppliedKey] = true

	for _, flag := range authFlags() {
		if hasFlag(app.Flags, flag) {
			continue
		}
		app.Flags = append(app.Flags, flag)
	}

	mergeAuthCommand(app)

	if !hasCommand(app.Commands, configCommand.Name) {
		app.Commands = append(app.Commands, configCommand)
	}

	if !hasCommand(app.Commands, downloadResultsCommand.Name) {
		app.Commands = append(app.Commands, downloadResultsCommand)
	}

	if !hasCommand(app.Commands, viewerCommand.Name) {
		app.Commands = append(app.Commands, viewerCommand)
	}

	addMergedInputFlags(app)
	addRunCommands(app)
	if !hasCommand(app.Commands, downloadStatusCommand.Name) {
		app.Commands = append(app.Commands, downloadStatusCommand)
	}
	customizeCommandTree(app)
}

func authFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "auth-issuer-url",
			Value:   authconfig.DefaultIssuerURL,
			Usage:   "OIDC issuer URL for OAuth login and bearer-token refresh",
			Sources: cli.EnvVars(authconfig.EnvAuthIssuerURL),
		},
		&cli.StringFlag{
			Name:    "auth-client-id",
			Value:   authconfig.DefaultClientID,
			Usage:   "OAuth client ID for the public client",
			Sources: cli.EnvVars(authconfig.EnvAuthClientID),
		},
		&cli.StringSliceFlag{
			Name:    "auth-scope",
			Usage:   "OAuth scope to request. Repeat the flag or use a comma-separated env value.",
			Sources: cli.EnvVars(authconfig.EnvAuthScope),
		},
		&cli.StringFlag{
			Name:    "auth-audience",
			Value:   authconfig.DefaultAudience,
			Usage:   "OAuth audience/resource to request during login",
			Sources: cli.EnvVars(authconfig.EnvAuthAudience),
		},
		&cli.StringFlag{
			Name:    "auth-authorization-url",
			Usage:   "Override the discovered authorization endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthAuthorizationURL),
		},
		&cli.StringFlag{
			Name:    "auth-token-url",
			Usage:   "Override the discovered token endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthTokenURL),
		},
		&cli.StringFlag{
			Name:    "auth-userinfo-url",
			Usage:   "Override the discovered userinfo endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthUserInfoURL),
		},
		&cli.StringFlag{
			Name:    "auth-revocation-url",
			Usage:   "Override the discovered revocation endpoint",
			Sources: cli.EnvVars(authconfig.EnvAuthRevocationURL),
		},
		&cli.StringFlag{
			Name:    "org",
			Usage:   "Local organization selection used by auth commands and future org context",
			Sources: cli.EnvVars(authconfig.EnvOrg),
		},
		&cli.IntFlag{
			Name:    "listen-port",
			Value:   authconfig.DefaultListenPort,
			Usage:   "Bind the OAuth loopback callback listener to this port",
			Sources: cli.EnvVars(authconfig.EnvListenPort),
		},
	}
}

func hasCommand(commands []*cli.Command, name string) bool {
	return findCommand(commands, name) != nil
}

func findCommand(commands []*cli.Command, name string) *cli.Command {
	for _, command := range commands {
		if command != nil && command.Name == name {
			return command
		}
	}
	return nil
}

func mergeAuthCommand(app *cli.Command) {
	existing := findCommand(app.Commands, authCommand.Name)
	if existing == nil {
		app.Commands = append(app.Commands, authCommand)
		return
	}

	// Stainless can generate API methods under `auth` (for example, `auth me`).
	// Keep those generated subcommands and layer our custom login/session commands alongside them.
	for _, customCommand := range authCommand.Commands {
		if !hasCommand(existing.Commands, customCommand.Name) {
			existing.Commands = append(existing.Commands, customCommand)
		}
	}
}

func hasFlag(flags []cli.Flag, candidate cli.Flag) bool {
	for _, flag := range flags {
		if slices.Equal(flag.Names(), candidate.Names()) {
			return true
		}
	}
	return false
}

const mergedInputUsage = "Structured request input. Merges fields into the top-level request body."

var mergedInputCommandPaths = [][]string{
	{"small-molecule:design", "estimate-cost"},
	{"small-molecule:design", "start"},
	{"small-molecule:library-screen", "estimate-cost"},
	{"small-molecule:library-screen", "start"},
	{"protein:design", "estimate-cost"},
	{"protein:design", "start"},
	{"protein:library-screen", "estimate-cost"},
	{"protein:library-screen", "start"},
}

var mergedInputExcludedBodyFields = map[string]struct{}{
	"idempotency_key": {},
	"workspace_id":    {},
}

func addMergedInputFlags(app *cli.Command) {
	for _, path := range mergedInputCommandPaths {
		cmd := findCommandByPath(app, path...)
		if cmd == nil {
			continue
		}
		addMergedInputFlag(cmd, mergedInputUsage)
	}
}

func addMergedInputFlag(cmd *cli.Command, usage string) {
	inputFlag := &requestflag.Flag[map[string]any]{
		Name:      "input",
		Usage:     usage,
		BodyMerge: true,
	}
	if hasFlag(cmd.Flags, inputFlag) {
		return
	}

	flagsToInsert := []cli.Flag{inputFlag}
	for _, innerFlag := range deriveMergedInputInnerFlags(cmd.Flags) {
		innerFlag.SetOuterFlag(inputFlag)
		flagsToInsert = append(flagsToInsert, innerFlag)
	}

	cmd.Flags = slices.Insert(cmd.Flags, 0, flagsToInsert...)
}

func deriveMergedInputInnerFlags(flags []cli.Flag) []requestflag.HasOuterFlag {
	derived := []requestflag.HasOuterFlag{}
	seen := map[string]struct{}{}

	for _, flag := range flags {
		inReq, ok := flag.(requestflag.InRequest)
		if !ok || inReq.GetBodyPath() == "" || inReq.IsBodyRoot() || inReq.IsBodyMerge() {
			continue
		}

		bodyPath := inReq.GetBodyPath()
		if _, excluded := mergedInputExcludedBodyFields[bodyPath]; excluded {
			continue
		}

		name := "input." + strings.ReplaceAll(bodyPath, "_", "-")
		if _, exists := seen[name]; exists {
			continue
		}

		innerFlag, ok := requestflag.NewDerivedInnerFlag(flag, name, bodyPath)
		if !ok {
			continue
		}
		derived = append(derived, innerFlag)
		seen[name] = struct{}{}
	}

	return derived
}

func findCommandByPath(root *cli.Command, path ...string) *cli.Command {
	cmd := root
	for _, segment := range path {
		var next *cli.Command
		for _, child := range cmd.Commands {
			if child != nil && child.Name == segment {
				next = child
				break
			}
		}
		if next == nil {
			return nil
		}
		cmd = next
	}
	return cmd
}

func customizeCommandTree(cmd *cli.Command) {
	if cmd == nil {
		return
	}

	supportsMergedInput := commandHasMergedInputFlag(cmd)
	if supportsMergedInput {
		maybeAnnotateMergedInputCommand(cmd)
	}

	for _, flag := range cmd.Flags {
		maybeCustomizeRootFlag(flag)
		maybeAnnotateRepeatableArrayFlag(flag)
		if supportsMergedInput {
			maybeAnnotateMergedInputFlag(flag)
		}
	}

	for _, child := range cmd.Commands {
		customizeCommandTree(child)
	}
}

func maybeCustomizeRootFlag(flag cli.Flag) {
	if canonicalFlagName(flag) != "transform" {
		return
	}
	setFlagStringField(flag, "Usage", transformUsage)
}

func maybeAnnotateRepeatableArrayFlag(flag cli.Flag) {
	inReq, ok := flag.(requestflag.InRequest)
	if !ok || inReq.GetBodyPath() == "" || !flagDefaultKindIs(flag, reflect.Slice) {
		return
	}

	flagName := canonicalFlagName(flag)
	bodyField := inReq.GetBodyPath()
	if !isSingularPluralRename(flagName, bodyField) {
		return
	}

	usage, ok := flagStringField(flag, "Usage")
	if !ok {
		return
	}

	note := fmt.Sprintf("Repeat --%s to add entries. When piping JSON or YAML, use the body field %s.", flagName, bodyField)
	if strings.Contains(usage, note) {
		return
	}
	if usage == "" {
		setFlagStringField(flag, "Usage", note)
		return
	}
	setFlagStringField(flag, "Usage", usage+" "+note)
}

func commandHasMergedInputFlag(cmd *cli.Command) bool {
	for _, flag := range cmd.Flags {
		if canonicalFlagName(flag) != "input" {
			continue
		}
		inReq, ok := flag.(requestflag.InRequest)
		if ok && inReq.IsBodyMerge() {
			return true
		}
	}
	return false
}

func maybeAnnotateMergedInputCommand(cmd *cli.Command) {
	if strings.Contains(cmd.Usage, mergedInputCommandNote) {
		return
	}
	if cmd.Usage == "" {
		cmd.Usage = mergedInputCommandNote
		return
	}
	cmd.Usage = cmd.Usage + " " + mergedInputCommandNote
}

func maybeAnnotateMergedInputFlag(flag cli.Flag) {
	flagName := canonicalFlagName(flag)
	if flagName == "" {
		return
	}

	usage, ok := flagStringField(flag, "Usage")
	if !ok {
		return
	}

	var note string
	if flagName == "input" {
		note = mergedInputFlagNote
	} else if inReq, ok := flag.(requestflag.InRequest); ok {
		if bodyPath := inReq.GetBodyPath(); bodyPath != "" && !inReq.IsBodyRoot() && !inReq.IsBodyMerge() {
			if _, reserved := mergedInputExcludedBodyFields[bodyPath]; reserved {
				note = mergedReservedFieldNote
			} else {
				note = mergedFieldOverrideNote
			}
		}
	}

	if note == "" || strings.Contains(usage, note) {
		return
	}
	if usage == "" {
		setFlagStringField(flag, "Usage", note)
		return
	}
	setFlagStringField(flag, "Usage", usage+" "+note)
}

func canonicalFlagName(flag cli.Flag) string {
	names := flag.Names()
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

func flagDefaultKindIs(flag cli.Flag, want reflect.Kind) bool {
	field, ok := flagField(flag, "Default")
	return ok && field.Kind() == want
}

func flagStringField(flag cli.Flag, name string) (string, bool) {
	field, ok := flagField(flag, name)
	if !ok || field.Kind() != reflect.String {
		return "", false
	}
	return field.String(), true
}

func setFlagStringField(flag cli.Flag, name, value string) bool {
	field, ok := flagField(flag, name)
	if !ok || field.Kind() != reflect.String || !field.CanSet() {
		return false
	}
	field.SetString(value)
	return true
}

func flagField(flag cli.Flag, name string) (reflect.Value, bool) {
	v := reflect.ValueOf(flag)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() {
		return reflect.Value{}, false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	field := v.FieldByName(name)
	if !field.IsValid() {
		return reflect.Value{}, false
	}
	return field, true
}

func isSingularPluralRename(flagName, bodyField string) bool {
	normalizedFlag := strings.ReplaceAll(flagName, "-", "_")
	return pluralize(normalizedFlag) == bodyField
}

func pluralize(s string) string {
	switch {
	case len(s) >= 2 && strings.HasSuffix(s, "y") && !isASCIIvowel(s[len(s)-2]):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "z"),
		strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

func isASCIIvowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}
