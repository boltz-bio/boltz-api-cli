// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/boltz-bio/boltz-api-cli/internal/apiquery"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
)

type runResourceKind string

const (
	runResourceKindPrediction runResourceKind = "prediction"
	runResourceKindAdme       runResourceKind = "adme"
	runResourceKindPipeline   runResourceKind = "pipeline"
)

type runResourceSpec struct {
	Path    []string
	RunType downloadRunType
	Kind    runResourceKind
	Start   func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error
}

type startedRemoteRun struct {
	ID          string
	WorkspaceID *string
}

type runCommandOptions struct {
	Name                *string
	RunDir              *string
	RootDir             string
	PollIntervalSeconds float64
	ProgressFormat      string
	DownloadMode        string
	DownloadModeSet     bool
	Workers             int
	Verbose             bool
}

const structureAndBindingRunInputUsage = "Structure and binding prediction input. Pass inline JSON/YAML, or use @json://... / @yaml://... for larger payload files. Include entities plus optional constraints, bonds, templates, binding settings, sample count, and model options. Keep --model, --idempotency-key, and --workspace-id as top-level flags."

var runResourceSpecs = []runResourceSpec{
	{
		Path:    []string{"predictions:structure-and-binding"},
		RunType: downloadRunTypePrediction,
		Kind:    runResourceKindPrediction,
		Start: func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error {
			_, err := client.Predictions.StructureAndBinding.Start(ctx, boltzapi.PredictionStructureAndBindingStartParams{}, options...)
			return err
		},
	},
	{
		Path:    []string{"predictions:adme"},
		RunType: downloadRunTypeAdme,
		Kind:    runResourceKindAdme,
		Start: func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error {
			_, err := client.Predictions.Adme.Start(ctx, boltzapi.PredictionAdmeStartParams{}, options...)
			return err
		},
	},
	{
		Path:    []string{"protein:design"},
		RunType: downloadRunTypeProteinDesign,
		Kind:    runResourceKindPipeline,
		Start: func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error {
			_, err := client.Protein.Design.Start(ctx, boltzapi.ProteinDesignStartParams{}, options...)
			return err
		},
	},
	{
		Path:    []string{"protein:library-screen"},
		RunType: downloadRunTypeProteinLibraryScreen,
		Kind:    runResourceKindPipeline,
		Start: func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error {
			_, err := client.Protein.LibraryScreen.Start(ctx, boltzapi.ProteinLibraryScreenStartParams{}, options...)
			return err
		},
	},
	{
		Path:    []string{"small-molecule:design"},
		RunType: downloadRunTypeSmallMoleculeDesign,
		Kind:    runResourceKindPipeline,
		Start: func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error {
			_, err := client.SmallMolecule.Design.Start(ctx, boltzapi.SmallMoleculeDesignStartParams{}, options...)
			return err
		},
	},
	{
		Path:    []string{"small-molecule:library-screen"},
		RunType: downloadRunTypeSmallMoleculeScreen,
		Kind:    runResourceKindPipeline,
		Start: func(ctx context.Context, client *boltzapi.Client, options []option.RequestOption) error {
			_, err := client.SmallMolecule.LibraryScreen.Start(ctx, boltzapi.SmallMoleculeLibraryScreenStartParams{}, options...)
			return err
		},
	},
}

func addRunCommands(app *cli.Command) {
	for _, spec := range runResourceSpecs {
		spec := spec
		parent := findCommandByPath(app, spec.Path...)
		if parent == nil || hasCommand(parent.Commands, "run") {
			continue
		}

		startCommand := findCommand(parent.Commands, "start")
		if startCommand == nil {
			continue
		}

		parent.Commands = insertCommandAfter(parent.Commands, "start", newRunCommand(spec, startCommand))
	}
}

func newRunCommand(spec runResourceSpec, startCommand *cli.Command) *cli.Command {
	flags := cloneCommandFlags(startCommand.Flags)
	annotateRunInputFlags(spec, flags)

	return &cli.Command{
		Name:            "run",
		Usage:           runCommandUsage(spec.Kind),
		Suggest:         true,
		Flags:           append(flags, runCommandFlags(spec.Kind)...),
		Action:          func(ctx context.Context, cmd *cli.Command) error { return handleResourceRun(ctx, cmd, spec) },
		HideHelpCommand: true,
	}
}

func annotateRunInputFlags(spec runResourceSpec, flags []cli.Flag) {
	if len(spec.Path) != 1 || spec.Path[0] != "predictions:structure-and-binding" {
		return
	}

	inputFlag := findFlagInList(flags, "input")
	if inputFlag == nil {
		return
	}

	usage, ok := flagStringField(inputFlag, "Usage")
	if ok && strings.TrimSpace(usage) == "" {
		setFlagStringField(inputFlag, "Usage", structureAndBindingRunInputUsage)
	}
}

func findFlagInList(flags []cli.Flag, name string) cli.Flag {
	for _, flag := range flags {
		if canonicalFlagName(flag) == name {
			return flag
		}
	}
	return nil
}

func runCommandUsage(kind runResourceKind) string {
	switch kind {
	case runResourceKindAdme:
		return "Start an ADME prediction, wait for completion, and persist results"
	case runResourceKindPipeline:
		return "Start a pipeline run, wait for completion, and download results"
	default:
		return "Start a prediction, wait for completion, and download results"
	}
}

func runCommandFlags(kind runResourceKind) []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "Local run directory name under --root-dir",
		},
		&cli.StringFlag{
			Name:  "run-dir",
			Usage: "Explicit local run directory path",
		},
		&cli.StringFlag{
			Name:  "root-dir",
			Usage: "Root directory for generated local run directories",
			Value: downloadResultsDefaultRootDir,
		},
		&cli.Float64Flag{
			Name:  "poll-interval-seconds",
			Usage: "Polling interval while waiting for remote results",
			Value: 5.0,
		},
		&cli.StringFlag{
			Name:  "progress-format",
			Usage: "Format for progress logs written to stderr (`jsonl` or `text`). `jsonl` is the default agent-friendly format.",
			Value: downloadProgressFormatJSONL,
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Print text progress logs to stderr when --progress-format text is selected",
		},
	}
	if kind == runResourceKindPipeline {
		flags = append(flags,
			&cli.StringFlag{
				Name:  "download-mode",
				Usage: "Pipeline artifact download mode (`everything` or `metadata_only`).",
				Value: downloadModeEverything,
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"num-workers", "num_workers"},
				Usage:   "Number of concurrent pipeline result archive downloads",
				Value:   downloadResultsDefaultWorkers,
			},
		)
	}
	return flags
}

func handleResourceRun(ctx context.Context, cmd *cli.Command, spec runResourceSpec) error {
	if unusedArgs := cmd.Args().Slice(); len(unusedArgs) > 0 {
		return fmt.Errorf("Unexpected extra arguments: %v", unusedArgs)
	}

	runOptions, err := parseRunCommandOptions(cmd, spec.Kind)
	if err != nil {
		return err
	}
	if err := ensureRunLocalTargetAvailable(runOptions); err != nil {
		return err
	}

	requestOptions, err := flagOptions(
		cmd,
		apiquery.NestedQueryFormatBrackets,
		apiquery.ArrayQueryFormatRepeat,
		ApplicationJSON,
		false,
	)
	if err != nil {
		return err
	}

	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	var startResponse []byte
	requestOptions = append(requestOptions, option.WithResponseBodyInto(&startResponse))
	if err := spec.Start(ctx, &client, requestOptions); err != nil {
		return err
	}

	started, err := parseStartedRemoteRun(startResponse, spec)
	if err != nil {
		return err
	}

	downloadSpec := downloadResultsSpec{
		ID:                  &started.ID,
		Name:                runOptions.Name,
		RunDir:              runOptions.RunDir,
		RootDir:             runOptions.RootDir,
		WorkspaceID:         runWorkspaceID(cmd, started.WorkspaceID),
		PollIntervalSeconds: runOptions.PollIntervalSeconds,
		ProgressFormat:      runOptions.ProgressFormat,
		DownloadMode:        runOptions.DownloadMode,
		DownloadModeSet:     runOptions.DownloadModeSet,
		Workers:             runOptions.Workers,
		Verbose:             runOptions.Verbose,
	}

	engine := downloadResultsEngine{
		client: &client,
		sink: &downloadResultsSink{
			verbose:        runOptions.Verbose,
			progressFormat: runOptions.ProgressFormat,
			writer:         commandErrorWriter(cmd),
		},
	}

	runDir, err := engine.download(ctx, downloadSpec)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(commandWriter(cmd), runDir)
	return err
}

func parseRunCommandOptions(cmd *cli.Command, kind runResourceKind) (runCommandOptions, error) {
	name := trimOptionalString(cmd.String("name"))
	runDir := trimOptionalString(cmd.String("run-dir"))
	rootDir := strings.TrimSpace(cmd.String("root-dir"))
	pollIntervalSeconds := cmd.Float64("poll-interval-seconds")
	progressFormat := normalizeDownloadProgressFormat(cmd.String("progress-format"))
	downloadMode := downloadModeEverything
	downloadModeSet := false
	workers := downloadResultsDefaultWorkers

	if kind == runResourceKindPipeline {
		downloadMode = normalizeDownloadMode(cmd.String("download-mode"))
		downloadModeSet = cmd.IsSet("download-mode")
		workers = cmd.Int("workers")
	}

	if pollIntervalSeconds < 0 {
		return runCommandOptions{}, errors.New("--poll-interval-seconds must be non-negative")
	}
	if !isSupportedDownloadProgressFormat(progressFormat) {
		return runCommandOptions{}, fmt.Errorf("--progress-format must be one of: %s", strings.Join([]string{downloadProgressFormatText, downloadProgressFormatJSONL}, ", "))
	}
	if !isSupportedDownloadMode(downloadMode) {
		return runCommandOptions{}, fmt.Errorf("--download-mode must be one of: %s", strings.Join(supportedDownloadModes(), ", "))
	}
	if workers < 1 {
		return runCommandOptions{}, errors.New("--workers must be at least 1")
	}
	if name != nil && runDir != nil {
		return runCommandOptions{}, errors.New("--name and --run-dir are mutually exclusive")
	}
	if runDir != nil && cmd.IsSet("root-dir") {
		return runCommandOptions{}, errors.New("--root-dir cannot be used with --run-dir")
	}
	if name != nil {
		if _, err := validateDownloadRunName(*name); err != nil {
			return runCommandOptions{}, err
		}
	}
	if runDir != nil && strings.TrimSpace(*runDir) == "" {
		return runCommandOptions{}, errors.New("--run-dir must not be empty")
	}

	return runCommandOptions{
		Name:                name,
		RunDir:              runDir,
		RootDir:             rootDir,
		PollIntervalSeconds: pollIntervalSeconds,
		ProgressFormat:      progressFormat,
		DownloadMode:        downloadMode,
		DownloadModeSet:     downloadModeSet,
		Workers:             workers,
		Verbose:             cmd.Bool("verbose"),
	}, nil
}

func ensureRunLocalTargetAvailable(options runCommandOptions) error {
	if options.Name == nil && options.RunDir == nil {
		return ensureRunRootDirAvailable(options.RootDir)
	}

	runDir, err := resolveDownloadRunDir(downloadResultsSpec{
		Name:    options.Name,
		RunDir:  options.RunDir,
		RootDir: options.RootDir,
	})
	if err != nil {
		return err
	}

	if info, err := os.Stat(runDir); err == nil && !info.IsDir() {
		return fmt.Errorf("Run path is not a directory: %s", runDir)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if _, err := os.Stat(downloadMetadataPath(runDir)); err == nil {
		return fmt.Errorf("Run directory %s already contains %s; use download-results to resume it or choose a different --name/--run-dir", runDir, downloadResultsMetadataFileName)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func ensureRunRootDirAvailable(rootDir string) error {
	resolvedRootDir, err := resolveDownloadRootDir(rootDir)
	if err != nil {
		return err
	}

	if info, err := os.Stat(resolvedRootDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("Run root path is not a directory: %s", resolvedRootDir)
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return os.MkdirAll(resolvedRootDir, 0o755)
}

func parseStartedRemoteRun(raw []byte, spec runResourceSpec) (startedRemoteRun, error) {
	obj := gjson.ParseBytes(raw)
	runID := strings.TrimSpace(obj.Get("id").String())
	if runID == "" {
		return startedRemoteRun{}, errors.New("Start response did not include an id")
	}

	runType, err := inferDownloadRunType(runID)
	if err != nil {
		return startedRemoteRun{}, err
	}
	if runType != spec.RunType {
		return startedRemoteRun{}, fmt.Errorf("Start response returned %s run ID %q for %s", runType, runID, strings.Join(spec.Path, " "))
	}

	return startedRemoteRun{
		ID:          runID,
		WorkspaceID: trimOptionalString(obj.Get("workspace_id").String()),
	}, nil
}

func runWorkspaceID(cmd *cli.Command, startedWorkspaceID *string) *string {
	if startedWorkspaceID != nil {
		return startedWorkspaceID
	}
	return trimOptionalString(cmd.String("workspace-id"))
}

func insertCommandAfter(commands []*cli.Command, after string, command *cli.Command) []*cli.Command {
	for i, child := range commands {
		if child != nil && child.Name == after {
			return slicesInsertCommand(commands, i+1, command)
		}
	}
	return append(commands, command)
}

func slicesInsertCommand(commands []*cli.Command, index int, command *cli.Command) []*cli.Command {
	if index < 0 || index > len(commands) {
		return append(commands, command)
	}
	commands = append(commands, nil)
	copy(commands[index+1:], commands[index:])
	commands[index] = command
	return commands
}

func cloneCommandFlags(flags []cli.Flag) []cli.Flag {
	cloned := make([]cli.Flag, 0, len(flags))
	clonedByOriginalName := map[string]cli.Flag{}
	innerFlags := []struct {
		cloned        requestflag.HasOuterFlag
		originalOuter cli.Flag
	}{}

	for _, flag := range flags {
		clonedFlag := cloneCommandFlag(flag)
		cloned = append(cloned, clonedFlag)
		if name := canonicalFlagName(flag); name != "" {
			clonedByOriginalName[name] = clonedFlag
		}
		if clonedInner, ok := clonedFlag.(requestflag.HasOuterFlag); ok {
			if originalInner, ok := flag.(requestflag.HasOuterFlag); ok {
				innerFlags = append(innerFlags, struct {
					cloned        requestflag.HasOuterFlag
					originalOuter cli.Flag
				}{
					cloned:        clonedInner,
					originalOuter: originalInner.GetOuterFlag(),
				})
			}
		}
	}

	for _, inner := range innerFlags {
		if name := canonicalFlagName(inner.originalOuter); name != "" {
			if outer := clonedByOriginalName[name]; outer != nil {
				inner.cloned.SetOuterFlag(outer)
			}
		}
	}

	return cloned
}

func cloneCommandFlag(flag cli.Flag) cli.Flag {
	value := reflect.ValueOf(flag)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return flag
	}

	cloned := reflect.New(value.Elem().Type())
	cloned.Elem().Set(value.Elem())
	if clonedFlag, ok := cloned.Interface().(cli.Flag); ok {
		return clonedFlag
	}
	return flag
}
