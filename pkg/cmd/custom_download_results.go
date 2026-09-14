// Custom CLI extension code. Not generated.
package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltz-bio/boltz-api-go"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/boltz-bio/boltz-api-go/packages/pagination"
	"github.com/boltz-bio/boltz-api-go/packages/param"
	"github.com/urfave/cli/v3"
)

const (
	downloadResultsDefaultRootDir     = "boltz-experiments"
	downloadResultsMetadataFileName   = ".boltz-run.json"
	downloadResultsRunFileName        = "run.json"
	downloadResultsResultMetadataName = "metadata.json"
	downloadResultsResultIndexName    = "index.jsonl"
	downloadResultsResultIndexLock    = "index.jsonl.lock"
	downloadResultsSchemaVersion      = 1
	downloadResultsPageLimit          = 1000
	downloadResultsDefaultWorkers     = 32
	downloadResultsSummaryIntervalSec = 30 * time.Second
	downloadResultsManifestLockWait   = 30 * time.Second
	downloadProgressFormatJSONL       = "jsonl"
	downloadProgressFormatText        = "text"
	downloadModeEverything            = "everything"
	downloadModeMetadataOnly          = "metadata_only"
)

var (
	downloadResultsRunTypePrefixes = []struct {
		prefix  string
		runType downloadRunType
	}{
		{prefix: "adme_pred_", runType: downloadRunTypeAdme},
		// Structure predictions now use the sab_pred prefix; keep pred_ as a legacy alias.
		{prefix: "sab_pred_", runType: downloadRunTypePrediction},
		{prefix: "pred_", runType: downloadRunTypePrediction},
		{prefix: "prot_des_", runType: downloadRunTypeProteinDesign},
		{prefix: "prot_scr_", runType: downloadRunTypeProteinLibraryScreen},
		{prefix: "sm_des_", runType: downloadRunTypeSmallMoleculeDesign},
		{prefix: "sm_scr_", runType: downloadRunTypeSmallMoleculeScreen},
	}
	downloadResultsUnsupportedPrefixes   = []string{"pres_", "psmp_"}
	downloadResultsAdjectives            = []string{"amber", "brisk", "calm", "clear", "keen", "lively", "lucid", "quiet", "rapid", "steady"}
	downloadResultsNouns                 = []string{"atom", "binder", "cluster", "enzyme", "helix", "ligand", "motif", "pocket", "sample", "target"}
	downloadResultsVerbs                 = []string{"aligns", "builds", "checks", "drifts", "folds", "guides", "mixes", "screens", "shifts", "tracks"}
	downloadResultsKnownMetadataFields   = map[string]struct{}{"schema_version": {}, "name": {}, "run_type": {}, "download_mode": {}, "request_fingerprint": {}, "idempotency_key": {}, "remote": {}, "pending": {}, "cursor_after_id": {}}
	downloadResultsSupportedArchiveTypes = []string{".tar.gz", ".tgz", ".tar", ".zip"}
)

var downloadResultsCommand = &cli.Command{
	Name:            "download-results",
	Usage:           "Download and resume result archives for predictions and pipelines",
	Suggest:         true,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "Prediction or pipeline run ID",
		},
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
		&cli.StringFlag{
			Name:  "workspace-id",
			Usage: "Workspace ID. Only used with admin API keys. Ignored (or validated) for workspace-scoped keys.",
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
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Print progress logs to stderr while downloading",
		},
	},
	Action: handleDownloadResults,
}

type downloadRunType string

const (
	downloadRunTypeAdme                  downloadRunType = "adme"
	downloadRunTypePrediction            downloadRunType = "prediction"
	downloadRunTypeProteinDesign         downloadRunType = "protein_design"
	downloadRunTypeProteinLibraryScreen  downloadRunType = "protein_library_screen"
	downloadRunTypeSmallMoleculeDesign   downloadRunType = "small_molecule_design"
	downloadRunTypeSmallMoleculeScreen   downloadRunType = "small_molecule_library_screen"
	downloadPendingKindPredictionArchive string          = "prediction_archive"
	downloadPendingKindResultPage        string          = "result_page"
)

type downloadResultsSpec struct {
	ID                  *string
	Name                *string
	RunDir              *string
	RootDir             string
	WorkspaceID         *string
	PollIntervalSeconds float64
	ProgressFormat      string
	DownloadMode        string
	DownloadModeSet     bool
	Workers             int
	Verbose             bool
}

type downloadRunMetadata struct {
	SchemaVersion      int                        `json:"schema_version"`
	Name               string                     `json:"name"`
	RunType            downloadRunType            `json:"run_type"`
	DownloadMode       string                     `json:"download_mode"`
	RequestFingerprint string                     `json:"request_fingerprint"`
	IdempotencyKey     string                     `json:"idempotency_key"`
	Remote             downloadRemoteState        `json:"remote"`
	Pending            *downloadPendingState      `json:"pending"`
	CursorAfterID      *string                    `json:"cursor_after_id"`
	ExtraFields        map[string]json.RawMessage `json:"-"`
}

type downloadRemoteState struct {
	RunID          *string `json:"run_id"`
	WorkspaceID    *string `json:"workspace_id"`
	Status         *string `json:"status"`
	StartedAt      *string `json:"started_at"`
	CompletedAt    *string `json:"completed_at"`
	StoppedAt      *string `json:"stopped_at"`
	LatestResultID *string `json:"latest_result_id"`
	ErrorCode      *string `json:"error_code"`
}

type downloadPendingState struct {
	Kind       string   `json:"kind"`
	AfterID    *string  `json:"after_id"`
	PageLastID *string  `json:"page_last_id"`
	ResultIDs  []string `json:"result_ids"`
}

type downloadStatusResponse struct {
	RunDir            string          `json:"run_dir"`
	Name              string          `json:"name"`
	RunType           downloadRunType `json:"run_type"`
	DownloadMode      string          `json:"download_mode"`
	RunID             *string         `json:"run_id,omitempty"`
	WorkspaceID       *string         `json:"workspace_id,omitempty"`
	Status            *string         `json:"status,omitempty"`
	Phase             string          `json:"phase"`
	Ready             bool            `json:"ready"`
	PendingKind       *string         `json:"pending_kind,omitempty"`
	PendingCount      int             `json:"pending_count"`
	PendingAfterID    *string         `json:"pending_after_id,omitempty"`
	PendingPageLastID *string         `json:"pending_page_last_id,omitempty"`
	PendingResultIDs  []string        `json:"pending_result_ids,omitempty"`
	CursorAfterID     *string         `json:"cursor_after_id,omitempty"`
	LatestResultID    *string         `json:"latest_result_id,omitempty"`
	StartedAt         *string         `json:"started_at,omitempty"`
	CompletedAt       *string         `json:"completed_at,omitempty"`
	StoppedAt         *string         `json:"stopped_at,omitempty"`
	ErrorCode         *string         `json:"error_code,omitempty"`
}

type downloadResultsEngine struct {
	client *boltzapi.Client
	sink   *downloadResultsSink
}

type downloadResultsSink struct {
	mu                   sync.Mutex
	verbose              bool
	progressFormat       string
	writer               io.Writer
	lastRunningSummaryAt time.Time
}

type downloadPredictionRunInfo struct {
	Status      string
	WorkspaceID *string
	StartedAt   *string
	CompletedAt *string
	ErrorCode   *string
	ArchiveURL  *string
	Run         map[string]any
}

type downloadPipelineRunInfo struct {
	Status         string
	WorkspaceID    *string
	StartedAt      *string
	CompletedAt    *string
	StoppedAt      *string
	LatestResultID *string
	ErrorCode      *string
	Run            map[string]any
}

type downloadPipelineResultInfo struct {
	ID         string
	ArchiveURL string
	Metadata   map[string]any
}

type downloadPipelinePage struct {
	Results []downloadPipelineResultInfo
	LastID  *string
}

type downloadPipelineAdapter struct {
	get         func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error)
	listResults func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error)
}

var downloadPipelineAdapters = map[downloadRunType]downloadPipelineAdapter{
	downloadRunTypeProteinDesign: {
		get: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzapi.ProteinDesignGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.Protein.Design.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.ID,
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
				response.RawJSON(),
			), nil
		},
		listResults: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzapi.ProteinDesignListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.Protein.Design.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzapi.ProteinDesignListResultsResponseUnion) downloadPipelineResultInfo {
				return newDownloadPipelineResultInfo(result.ID, result.Artifacts.Archive.URL, result.RawJSON())
			}), nil
		},
	},
	downloadRunTypeProteinLibraryScreen: {
		get: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzapi.ProteinLibraryScreenGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.Protein.LibraryScreen.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.ID,
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
				response.RawJSON(),
			), nil
		},
		listResults: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzapi.ProteinLibraryScreenListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.Protein.LibraryScreen.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzapi.ProteinLibraryScreenListResultsResponse) downloadPipelineResultInfo {
				return newDownloadPipelineResultInfo(result.ID, result.Artifacts.Archive.URL, result.RawJSON())
			}), nil
		},
	},
	downloadRunTypeSmallMoleculeDesign: {
		get: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzapi.SmallMoleculeDesignGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.SmallMolecule.Design.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.ID,
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
				response.RawJSON(),
			), nil
		},
		listResults: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzapi.SmallMoleculeDesignListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.SmallMolecule.Design.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzapi.SmallMoleculeDesignListResultsResponse) downloadPipelineResultInfo {
				return newDownloadPipelineResultInfo(result.ID, result.Artifacts.Archive.URL, result.RawJSON())
			}), nil
		},
	},
	downloadRunTypeSmallMoleculeScreen: {
		get: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
			params := boltzapi.SmallMoleculeLibraryScreenGetParams{}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			response, err := client.SmallMolecule.LibraryScreen.Get(ctx, runID, params)
			if err != nil {
				return downloadPipelineRunInfo{}, err
			}
			return newDownloadPipelineRunInfo(
				string(response.Status),
				response.ID,
				response.WorkspaceID,
				response.StartedAt,
				response.CompletedAt,
				response.StoppedAt,
				response.Progress.LatestResultID,
				response.Error.Code,
				response.RawJSON(),
			), nil
		},
		listResults: func(ctx context.Context, client *boltzapi.Client, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
			params := boltzapi.SmallMoleculeLibraryScreenListResultsParams{Limit: param.NewOpt(int64(downloadResultsPageLimit))}
			setOptionalStringOpt(&params.WorkspaceID, workspaceID)
			setOptionalStringOpt(&params.AfterID, afterID)
			response, err := client.SmallMolecule.LibraryScreen.ListResults(ctx, runID, params)
			if err != nil {
				return downloadPipelinePage{}, err
			}
			return normalizePipelinePage(response, func(result boltzapi.SmallMoleculeLibraryScreenListResultsResponse) downloadPipelineResultInfo {
				return newDownloadPipelineResultInfo(result.ID, result.Artifacts.Archive.URL, result.RawJSON())
			}), nil
		},
	},
}

func handleDownloadResults(ctx context.Context, cmd *cli.Command) error {
	spec, err := parseDownloadResultsSpec(cmd)
	if err != nil {
		return err
	}

	client := boltzapi.NewClient(getDefaultRequestOptions(cmd)...)
	engine := downloadResultsEngine{
		client: &client,
		sink: &downloadResultsSink{
			verbose:        spec.Verbose,
			progressFormat: spec.ProgressFormat,
			writer:         commandErrorWriter(cmd),
		},
	}

	runDir, err := engine.download(ctx, spec)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(commandWriter(cmd), runDir)
	return err
}

func parseDownloadResultsSpec(cmd *cli.Command) (downloadResultsSpec, error) {
	unusedArgs := cmd.Args().Slice()
	id := trimOptionalString(cmd.String("id"))
	if len(unusedArgs) > 0 {
		return downloadResultsSpec{}, fmt.Errorf("Unexpected extra arguments: %v. Use --id for the run ID.", unusedArgs)
	}

	name := trimOptionalString(cmd.String("name"))
	runDir := trimOptionalString(cmd.String("run-dir"))
	rootDir := strings.TrimSpace(cmd.String("root-dir"))
	workspaceID := trimOptionalString(cmd.String("workspace-id"))
	pollIntervalSeconds := cmd.Float64("poll-interval-seconds")
	progressFormat := normalizeDownloadProgressFormat(cmd.String("progress-format"))
	downloadMode := normalizeDownloadMode(cmd.String("download-mode"))
	workers := cmd.Int("workers")
	if pollIntervalSeconds < 0 {
		return downloadResultsSpec{}, errors.New("--poll-interval-seconds must be non-negative")
	}
	if !isSupportedDownloadProgressFormat(progressFormat) {
		return downloadResultsSpec{}, fmt.Errorf("--progress-format must be one of: %s", strings.Join([]string{downloadProgressFormatText, downloadProgressFormatJSONL}, ", "))
	}
	if !isSupportedDownloadMode(downloadMode) {
		return downloadResultsSpec{}, fmt.Errorf("--download-mode must be one of: %s", strings.Join(supportedDownloadModes(), ", "))
	}
	if workers < 1 {
		return downloadResultsSpec{}, errors.New("--workers must be at least 1")
	}
	if name != nil && runDir != nil {
		return downloadResultsSpec{}, errors.New("--name and --run-dir are mutually exclusive")
	}
	if runDir != nil && cmd.IsSet("root-dir") {
		return downloadResultsSpec{}, errors.New("--root-dir cannot be used with --run-dir")
	}
	if name != nil {
		if _, err := validateDownloadRunName(*name); err != nil {
			return downloadResultsSpec{}, err
		}
	}
	if runDir != nil && strings.TrimSpace(*runDir) == "" {
		return downloadResultsSpec{}, errors.New("--run-dir must not be empty")
	}
	if id != nil {
		runType, err := inferDownloadRunType(*id)
		if err != nil {
			return downloadResultsSpec{}, err
		}
		if !isPipelineDownloadRunType(runType) && downloadMode != downloadModeEverything {
			return downloadResultsSpec{}, errors.New("--download-mode only supports pipeline run IDs")
		}
	}

	return downloadResultsSpec{
		ID:                  id,
		Name:                name,
		RunDir:              runDir,
		RootDir:             rootDir,
		WorkspaceID:         workspaceID,
		PollIntervalSeconds: pollIntervalSeconds,
		ProgressFormat:      progressFormat,
		DownloadMode:        downloadMode,
		DownloadModeSet:     cmd.IsSet("download-mode"),
		Workers:             workers,
		Verbose:             cmd.Bool("verbose"),
	}, nil
}

func (e *downloadResultsEngine) download(ctx context.Context, spec downloadResultsSpec) (string, error) {
	runDir, metadata, err := prepareDownloadRun(spec)
	if err != nil {
		return "", err
	}

	switch metadata.RunType {
	case downloadRunTypePrediction:
		if err := e.waitForPrediction(ctx, runDir, &metadata, spec.PollIntervalSeconds); err != nil {
			return "", err
		}
	case downloadRunTypeAdme:
		if err := e.waitForAdmePrediction(ctx, runDir, &metadata, spec.PollIntervalSeconds); err != nil {
			return "", err
		}
	default:
		if err := e.waitForPipeline(ctx, runDir, &metadata, spec.PollIntervalSeconds, spec.Workers); err != nil {
			return "", err
		}
	}

	return runDir, nil
}

func prepareDownloadRun(spec downloadResultsSpec) (string, downloadRunMetadata, error) {
	runDir, err := resolveDownloadRunDir(spec)
	if err != nil {
		return "", downloadRunMetadata{}, err
	}

	metadataPath := downloadMetadataPath(runDir)
	if _, err := os.Stat(metadataPath); err == nil {
		metadata, err := loadDownloadMetadata(runDir)
		if err != nil {
			return "", downloadRunMetadata{}, err
		}
		if err := reconcileMetadataWithSpec(runDir, &metadata, spec); err != nil {
			return "", downloadRunMetadata{}, err
		}
		return runDir, metadata, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", downloadRunMetadata{}, err
	}

	if spec.ID == nil {
		return "", downloadRunMetadata{}, fmt.Errorf("No local metadata exists for %s. Provide --id to create a new download run.", runDir)
	}

	if err := ensureDownloadDirectoryReady(runDir); err != nil {
		return "", downloadRunMetadata{}, err
	}

	runType, err := inferDownloadRunType(*spec.ID)
	if err != nil {
		return "", downloadRunMetadata{}, err
	}

	metadata := newDownloadRunMetadata(filepath.Base(runDir), runType, *spec.ID, spec.WorkspaceID, spec.DownloadMode)
	if err := saveDownloadMetadata(runDir, metadata); err != nil {
		return "", downloadRunMetadata{}, err
	}
	return runDir, metadata, nil
}

func resolveDownloadRunDir(spec downloadResultsSpec) (string, error) {
	if spec.RunDir != nil {
		return resolveDownloadPath(*spec.RunDir)
	}

	rootDir, err := resolveDownloadRootDir(spec.RootDir)
	if err != nil {
		return "", err
	}

	if spec.Name != nil {
		name, err := validateDownloadRunName(*spec.Name)
		if err != nil {
			return "", err
		}
		return filepath.Join(rootDir, name), nil
	}

	if spec.ID == nil {
		return "", errors.New("Either local metadata must already exist or --id must be provided")
	}

	return filepath.Join(rootDir, deterministicDownloadRunName(*spec.ID)), nil
}

func resolveDownloadRootDir(rootDir string) (string, error) {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = downloadResultsDefaultRootDir
	}
	return resolveDownloadPath(rootDir)
}

func resolveDownloadPath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("path must not be empty")
	}
	if strings.HasPrefix(trimmed, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		switch trimmed {
		case "~":
			trimmed = home
		default:
			if strings.HasPrefix(trimmed, "~/") || strings.HasPrefix(trimmed, `~\`) {
				trimmed = filepath.Join(home, trimmed[2:])
			}
		}
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed), nil
	}
	return filepath.Abs(trimmed)
}

func normalizeDownloadProgressFormat(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return downloadProgressFormatJSONL
	}
	return normalized
}

func isSupportedDownloadProgressFormat(value string) bool {
	return value == downloadProgressFormatText || value == downloadProgressFormatJSONL
}

func normalizeDownloadMode(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	if normalized == "" {
		return downloadModeEverything
	}
	if normalized == "all" {
		return downloadModeEverything
	}
	return normalized
}

func isSupportedDownloadMode(value string) bool {
	switch value {
	case downloadModeEverything, downloadModeMetadataOnly:
		return true
	default:
		return false
	}
}

func supportedDownloadModes() []string {
	return []string{downloadModeEverything, downloadModeMetadataOnly}
}

func validateDownloadRunName(name string) (string, error) {
	stripped := strings.TrimSpace(name)
	if stripped == "" {
		return "", errors.New("`name` must not be empty")
	}
	if stripped == "." || stripped == ".." {
		return "", errors.New("`name` must not be '.' or '..'")
	}
	if filepath.Base(stripped) != stripped {
		return "", errors.New("`name` must be a single directory name, not a path")
	}
	return stripped, nil
}

func buildDownloadStatusResponse(runDir string, metadata downloadRunMetadata) downloadStatusResponse {
	response := downloadStatusResponse{
		RunDir:         runDir,
		Name:           metadata.Name,
		RunType:        metadata.RunType,
		DownloadMode:   metadata.DownloadMode,
		RunID:          cloneString(metadata.Remote.RunID),
		WorkspaceID:    cloneString(metadata.Remote.WorkspaceID),
		Status:         cloneString(metadata.Remote.Status),
		Phase:          downloadRunPhase(metadata),
		Ready:          downloadRunReady(metadata),
		PendingCount:   0,
		CursorAfterID:  cloneString(metadata.CursorAfterID),
		LatestResultID: cloneString(metadata.Remote.LatestResultID),
		StartedAt:      cloneString(metadata.Remote.StartedAt),
		CompletedAt:    cloneString(metadata.Remote.CompletedAt),
		StoppedAt:      cloneString(metadata.Remote.StoppedAt),
		ErrorCode:      cloneString(metadata.Remote.ErrorCode),
	}
	if metadata.Pending != nil {
		response.PendingKind = cloneString(&metadata.Pending.Kind)
		response.PendingCount = len(metadata.Pending.ResultIDs)
		response.PendingAfterID = cloneString(metadata.Pending.AfterID)
		response.PendingPageLastID = cloneString(metadata.Pending.PageLastID)
		response.PendingResultIDs = cloneStringSlice(metadata.Pending.ResultIDs)
	}
	return response
}

func downloadRunPhase(metadata downloadRunMetadata) string {
	if metadata.Pending != nil {
		switch metadata.Pending.Kind {
		case downloadPendingKindPredictionArchive:
			return "materializing_outputs"
		case downloadPendingKindResultPage:
			return "materializing_results"
		default:
			return metadata.Pending.Kind
		}
	}

	switch derefString(metadata.Remote.Status) {
	case "":
		return "created"
	case "pending", "running":
		return "waiting_remote"
	case "failed":
		return "failed"
	case "succeeded":
		return "ready"
	case "stopped":
		return "stopped"
	default:
		return derefString(metadata.Remote.Status)
	}
}

func downloadRunReady(metadata downloadRunMetadata) bool {
	if metadata.Pending != nil {
		return false
	}
	status := derefString(metadata.Remote.Status)
	return status == "succeeded" || status == "stopped"
}

func deterministicDownloadRunName(runID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(runID)))
	return fmt.Sprintf(
		"%s-%s-%s-%s",
		deterministicDownloadWord(downloadResultsAdjectives, sum[0]),
		deterministicDownloadWord(downloadResultsNouns, sum[1]),
		deterministicDownloadWord(downloadResultsVerbs, sum[2]),
		hex.EncodeToString(sum[3:6]),
	)
}

func deterministicDownloadWord(values []string, seed byte) string {
	if len(values) == 0 {
		return "run"
	}
	return values[int(seed)%len(values)]
}

func randomHex(numBytes int) string {
	if numBytes <= 0 {
		return ""
	}
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		fallback := make([]byte, numBytes)
		for i := range fallback {
			fallback[i] = byte((time.Now().UnixNano() >> (i * 8)) & 0xff)
		}
		buf = fallback
	}
	return hex.EncodeToString(buf)
}

func ensureDownloadDirectoryReady(runDir string) error {
	info, err := os.Stat(runDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("Run path is not a directory: %s", runDir)
		}
		if _, err := os.Stat(downloadMetadataPath(runDir)); err == nil {
			return nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.MkdirAll(runDir, 0o755)
}

func reconcileMetadataWithSpec(runDir string, metadata *downloadRunMetadata, spec downloadResultsSpec) error {
	metadata.DownloadMode = normalizeDownloadMode(metadata.DownloadMode)
	if !isSupportedDownloadMode(metadata.DownloadMode) {
		return fmt.Errorf("Run directory %s has unsupported download mode %q", runDir, metadata.DownloadMode)
	}
	if spec.DownloadModeSet && metadata.DownloadMode != spec.DownloadMode {
		return fmt.Errorf("Run directory %s uses download mode %s, not %s", runDir, metadata.DownloadMode, spec.DownloadMode)
	}
	if spec.ID != nil {
		if metadata.Remote.RunID == nil || *metadata.Remote.RunID != *spec.ID {
			return fmt.Errorf("Run directory %s belongs to remote run %s, not %s", runDir, derefString(metadata.Remote.RunID), *spec.ID)
		}
		inferredType, err := inferDownloadRunType(*spec.ID)
		if err != nil {
			return err
		}
		if metadata.RunType != inferredType {
			return fmt.Errorf("Run directory %s belongs to %s, not %s", runDir, metadata.RunType, inferredType)
		}
		if !isPipelineDownloadRunType(metadata.RunType) && metadata.DownloadMode != downloadModeEverything {
			return fmt.Errorf("Run directory %s belongs to %s and cannot use download mode %s", runDir, metadata.RunType, metadata.DownloadMode)
		}
	}
	if spec.WorkspaceID != nil {
		if metadata.Remote.WorkspaceID == nil {
			metadata.Remote.WorkspaceID = cloneString(spec.WorkspaceID)
			metadata.RequestFingerprint = fingerprintDownloadRun(metadata.RunType, derefString(metadata.Remote.RunID), metadata.Remote.WorkspaceID)
			return saveDownloadMetadata(runDir, *metadata)
		}
		if *metadata.Remote.WorkspaceID != *spec.WorkspaceID {
			return fmt.Errorf("Run directory %s belongs to workspace %s, not %s", runDir, *metadata.Remote.WorkspaceID, *spec.WorkspaceID)
		}
	}
	if metadata.SchemaVersion != downloadResultsSchemaVersion {
		return fmt.Errorf("Unsupported experiments metadata schema version: %d", metadata.SchemaVersion)
	}
	return nil
}

func newDownloadRunMetadata(name string, runType downloadRunType, runID string, workspaceID *string, downloadMode string) downloadRunMetadata {
	downloadMode = normalizeDownloadMode(downloadMode)
	if !isSupportedDownloadMode(downloadMode) {
		downloadMode = downloadModeEverything
	}
	return downloadRunMetadata{
		SchemaVersion:      downloadResultsSchemaVersion,
		Name:               name,
		RunType:            runType,
		DownloadMode:       downloadMode,
		RequestFingerprint: fingerprintDownloadRun(runType, runID, workspaceID),
		IdempotencyKey:     "download_" + randomHex(16),
		Remote: downloadRemoteState{
			RunID:       cloneString(&runID),
			WorkspaceID: cloneString(workspaceID),
		},
		Pending:       nil,
		CursorAfterID: nil,
	}
}

func fingerprintDownloadRun(runType downloadRunType, runID string, workspaceID *string) string {
	payload := struct {
		RunType     downloadRunType `json:"run_type"`
		RunID       string          `json:"run_id"`
		WorkspaceID *string         `json:"workspace_id"`
	}{
		RunType:     runType,
		RunID:       runID,
		WorkspaceID: cloneString(workspaceID),
	}
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func (m *downloadRunMetadata) UnmarshalJSON(data []byte) error {
	type alias downloadRunMetadata
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key := range downloadResultsKnownMetadataFields {
		delete(raw, key)
	}

	*m = downloadRunMetadata(decoded)
	if len(raw) > 0 {
		m.ExtraFields = raw
	} else {
		m.ExtraFields = nil
	}
	return nil
}

func (m downloadRunMetadata) MarshalJSON() ([]byte, error) {
	type alias downloadRunMetadata
	known, err := json.Marshal(alias(m))
	if err != nil {
		return nil, err
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(known, &merged); err != nil {
		return nil, err
	}
	for key, value := range m.ExtraFields {
		if _, exists := merged[key]; !exists {
			merged[key] = value
		}
	}
	return json.Marshal(merged)
}

func loadDownloadMetadata(runDir string) (downloadRunMetadata, error) {
	raw, err := os.ReadFile(downloadMetadataPath(runDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return downloadRunMetadata{}, fmt.Errorf("Run metadata does not exist: %s", downloadMetadataPath(runDir))
		}
		return downloadRunMetadata{}, err
	}
	var metadata downloadRunMetadata
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata: %w", err)
	}
	if metadata.SchemaVersion != downloadResultsSchemaVersion {
		return downloadRunMetadata{}, fmt.Errorf("Unsupported experiments metadata schema version: %d", metadata.SchemaVersion)
	}
	if metadata.Name == "" {
		return downloadRunMetadata{}, errors.New("Invalid run metadata.name: required")
	}
	if !isSupportedDownloadRunType(metadata.RunType) {
		return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata.run_type: %q", metadata.RunType)
	}
	metadata.DownloadMode = normalizeDownloadMode(metadata.DownloadMode)
	if !isSupportedDownloadMode(metadata.DownloadMode) {
		return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata.download_mode: %q", metadata.DownloadMode)
	}
	if metadata.Pending != nil {
		if metadata.Pending.Kind != downloadPendingKindPredictionArchive && metadata.Pending.Kind != downloadPendingKindResultPage {
			return downloadRunMetadata{}, fmt.Errorf("Invalid run metadata.pending.kind: %q", metadata.Pending.Kind)
		}
	}
	return metadata, nil
}

func saveDownloadMetadata(runDir string, metadata downloadRunMetadata) error {
	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	path := downloadMetadataPath(runDir)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func saveDownloadRunMetadata(runDir string, run map[string]any) error {
	if len(run) == 0 {
		return nil
	}
	return writeDownloadJSONFile(filepath.Join(runDir, downloadResultsRunFileName), run)
}

func downloadMetadataPath(runDir string) string {
	return filepath.Join(runDir, downloadResultsMetadataFileName)
}

func isSupportedDownloadRunType(runType downloadRunType) bool {
	if runType == downloadRunTypePrediction || runType == downloadRunTypeAdme {
		return true
	}
	return isPipelineDownloadRunType(runType)
}

func isPipelineDownloadRunType(runType downloadRunType) bool {
	_, ok := downloadPipelineAdapters[runType]
	return ok
}

func inferDownloadRunType(id string) (downloadRunType, error) {
	for _, spec := range downloadResultsRunTypePrefixes {
		if strings.HasPrefix(id, spec.prefix) {
			return spec.runType, nil
		}
	}
	for _, prefix := range downloadResultsUnsupportedPrefixes {
		if strings.HasPrefix(id, prefix) {
			return "", fmt.Errorf("Unsupported ID prefix %q. download-results only supports prediction and pipeline run IDs.", strings.TrimSuffix(prefix, "_"))
		}
	}
	prefix := id
	if idx := strings.Index(prefix, "_"); idx >= 0 {
		prefix = prefix[:idx]
	}
	return "", fmt.Errorf("Could not infer download-results run type from ID prefix %q", prefix)
}

func (e *downloadResultsEngine) waitForPrediction(ctx context.Context, runDir string, metadata *downloadRunMetadata, pollIntervalSeconds float64) error {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return fmt.Errorf("Run %s has no confirmed remote run ID", runDir)
	}
	e.sink.info("waiting", fmt.Sprintf("Waiting for prediction in %s", runDir), runDir, metadata, nil)

	for {
		response, err := e.getPrediction(ctx, runID, metadata.Remote.WorkspaceID)
		if err != nil {
			return err
		}
		previousStatus := derefString(metadata.Remote.Status)
		updateRunMetadata(metadata, response.Status, response.WorkspaceID, response.StartedAt, response.CompletedAt, nil, nil, response.ErrorCode)
		if err := saveDownloadMetadata(runDir, *metadata); err != nil {
			return err
		}
		if err := saveDownloadRunMetadata(runDir, response.Run); err != nil {
			return err
		}
		if response.Status != previousStatus {
			e.sink.info("status", fmt.Sprintf("Prediction status: %s", response.Status), runDir, metadata, nil)
		}

		switch response.Status {
		case "pending", "running":
			e.sink.maybeRunningSummary("running", "Prediction still running...", runDir, metadata, nil)
			if err := sleepWithContext(ctx, pollIntervalSeconds); err != nil {
				return err
			}
			continue
		case "failed":
			failed := failureForMetadata(*metadata)
			e.sink.info("failed", failed.Error(), runDir, metadata, nil)
			return failed
		case "succeeded":
			if response.ArchiveURL == nil || strings.TrimSpace(*response.ArchiveURL) == "" {
				return errors.New("Prediction succeeded but did not return an archive URL")
			}
			archivePath, extractedDir, err := materializationPaths(filepath.Join(runDir, "outputs"), *response.ArchiveURL)
			if err != nil {
				return err
			}
			if !isDownloadMaterialized(archivePath, extractedDir) {
				metadata.Pending = &downloadPendingState{Kind: downloadPendingKindPredictionArchive}
				if err := saveDownloadMetadata(runDir, *metadata); err != nil {
					return err
				}
				if err := e.materializeArchive(ctx, runDir, metadata, *response.ArchiveURL, archivePath, extractedDir, nil); err != nil {
					return err
				}
				metadata.Pending = nil
				if err := saveDownloadMetadata(runDir, *metadata); err != nil {
					return err
				}
			}
			if err := prefixPredictionStructureFile(runDir, runID); err != nil {
				return err
			}
			e.sink.info("ready", fmt.Sprintf("Prediction ready in %s", runDir), runDir, metadata, nil)
			return nil
		default:
			return fmt.Errorf("Unsupported prediction status %q", response.Status)
		}
	}
}

func (e *downloadResultsEngine) waitForAdmePrediction(ctx context.Context, runDir string, metadata *downloadRunMetadata, pollIntervalSeconds float64) error {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return fmt.Errorf("Run %s has no confirmed remote run ID", runDir)
	}
	e.sink.info("waiting", fmt.Sprintf("Waiting for ADME prediction in %s", runDir), runDir, metadata, nil)

	for {
		response, err := e.getAdmePrediction(ctx, runID, metadata.Remote.WorkspaceID)
		if err != nil {
			return err
		}
		previousStatus := derefString(metadata.Remote.Status)
		updateRunMetadata(metadata, response.Status, response.WorkspaceID, response.StartedAt, response.CompletedAt, nil, nil, response.ErrorCode)
		if err := saveDownloadMetadata(runDir, *metadata); err != nil {
			return err
		}
		if err := saveDownloadRunMetadata(runDir, response.Run); err != nil {
			return err
		}
		if response.Status != previousStatus {
			e.sink.info("status", fmt.Sprintf("ADME prediction status: %s", response.Status), runDir, metadata, nil)
		}

		switch response.Status {
		case "pending", "running":
			e.sink.maybeRunningSummary("running", "ADME prediction still running...", runDir, metadata, nil)
			if err := sleepWithContext(ctx, pollIntervalSeconds); err != nil {
				return err
			}
			continue
		case "failed":
			failed := failureForMetadata(*metadata)
			e.sink.info("failed", failed.Error(), runDir, metadata, nil)
			return failed
		case "succeeded":
			e.sink.info("ready", fmt.Sprintf("ADME prediction ready in %s", runDir), runDir, metadata, nil)
			return nil
		default:
			return fmt.Errorf("Unsupported ADME prediction status %q", response.Status)
		}
	}
}

func (e *downloadResultsEngine) waitForPipeline(ctx context.Context, runDir string, metadata *downloadRunMetadata, pollIntervalSeconds float64, workers int) error {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return fmt.Errorf("Run %s has no confirmed remote run ID", runDir)
	}
	e.sink.info("waiting", fmt.Sprintf("Waiting for %s in %s", strings.ReplaceAll(string(metadata.RunType), "_", " "), runDir), runDir, metadata, nil)
	manifest := newPipelineResultManifestAppender(runDir, metadata.RunType)

	for {
		response, err := e.getPipeline(ctx, metadata.RunType, runID, metadata.Remote.WorkspaceID)
		if err != nil {
			return err
		}
		previousStatus := derefString(metadata.Remote.Status)
		updateRunMetadata(metadata, response.Status, response.WorkspaceID, response.StartedAt, response.CompletedAt, response.StoppedAt, response.LatestResultID, response.ErrorCode)
		if err := saveDownloadMetadata(runDir, *metadata); err != nil {
			return err
		}
		if err := saveDownloadRunMetadata(runDir, response.Run); err != nil {
			return err
		}
		if response.Status != previousStatus {
			e.sink.info("status", fmt.Sprintf("Pipeline status: %s", response.Status), runDir, metadata, nil)
		}

		madeProgress := false
		for {
			if metadata.Pending != nil && metadata.Pending.Kind == downloadPendingKindResultPage {
				if err := e.drainPendingPipelinePage(ctx, runDir, metadata, manifest, workers); err != nil {
					return err
				}
				madeProgress = true
				continue
			}
			discovered, err := e.discoverNextPipelinePage(ctx, runDir, metadata)
			if err != nil {
				return err
			}
			if discovered {
				if err := saveDownloadMetadata(runDir, *metadata); err != nil {
					return err
				}
				madeProgress = true
				continue
			}
			break
		}

		switch response.Status {
		case "pending", "running":
			if !madeProgress {
				e.sink.maybeRunningSummary("running", pipelineRunningSummary(metadata.Remote.LatestResultID), runDir, metadata, nil)
				if err := sleepWithContext(ctx, pollIntervalSeconds); err != nil {
					return err
				}
			}
			continue
		case "failed":
			failed := failureForMetadata(*metadata)
			e.sink.info("failed", failed.Error(), runDir, metadata, nil)
			return failed
		case "succeeded", "stopped":
			e.sink.info("ready", fmt.Sprintf("Pipeline ready in %s", runDir), runDir, metadata, nil)
			return nil
		default:
			return fmt.Errorf("Unsupported pipeline status %q", response.Status)
		}
	}
}

func (e *downloadResultsEngine) discoverNextPipelinePage(ctx context.Context, runDir string, metadata *downloadRunMetadata) (bool, error) {
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return false, errors.New("Pipeline metadata is missing a remote run ID")
	}
	page, err := e.listPipelineResults(ctx, metadata.RunType, runID, metadata.Remote.WorkspaceID, metadata.CursorAfterID)
	if err != nil {
		return false, err
	}
	if len(page.Results) == 0 {
		return false, nil
	}

	pageLastID := cloneString(page.LastID)
	if pageLastID == nil && len(page.Results) > 0 {
		lastID := page.Results[len(page.Results)-1].ID
		pageLastID = &lastID
	}

	resultIDs := make([]string, 0, len(page.Results))
	for _, result := range page.Results {
		resultIDs = append(resultIDs, result.ID)
	}

	metadata.Pending = &downloadPendingState{
		Kind:       downloadPendingKindResultPage,
		AfterID:    cloneString(metadata.CursorAfterID),
		PageLastID: pageLastID,
		ResultIDs:  resultIDs,
	}
	e.sink.info("page_discovered", fmt.Sprintf("Queued %d results for download", len(resultIDs)), runDir, metadata, nil)
	return true, nil
}

func (e *downloadResultsEngine) drainPendingPipelinePage(ctx context.Context, runDir string, metadata *downloadRunMetadata, manifest *pipelineResultManifestAppender, workers int) error {
	if metadata.Pending == nil || metadata.Pending.Kind != downloadPendingKindResultPage {
		return errors.New("Pipeline metadata does not contain a pending page")
	}
	runID := derefString(metadata.Remote.RunID)
	if runID == "" {
		return errors.New("Pipeline metadata is missing a remote run ID")
	}

	page, err := e.listPipelineResults(ctx, metadata.RunType, runID, metadata.Remote.WorkspaceID, metadata.Pending.AfterID)
	if err != nil {
		return err
	}
	results := make(map[string]downloadPipelineResultInfo, len(page.Results))
	for _, result := range page.Results {
		results[result.ID] = result
	}

	pendingResults := make([]downloadPipelineResultInfo, 0, len(metadata.Pending.ResultIDs))
	for _, resultID := range metadata.Pending.ResultIDs {
		result, ok := results[resultID]
		if !ok {
			return fmt.Errorf("Unable to resume result page; result %s is no longer present", resultID)
		}
		pendingResults = append(pendingResults, result)
	}

	switch metadata.DownloadMode {
	case downloadModeMetadataOnly:
		for _, result := range pendingResults {
			if err := manifest.appendMetadata(result.Metadata, result.ID); err != nil {
				return err
			}
			metadata.Pending.ResultIDs = metadata.Pending.ResultIDs[1:]
			if err := saveDownloadMetadata(runDir, *metadata); err != nil {
				return err
			}
		}
	case downloadModeEverything:
		materialized, materializeErr := e.materializePendingPipelineResults(ctx, runDir, metadata, pendingResults, workers)
		appendable := orderedMaterializedPrefix(pendingResults, materialized)
		appendedIDs := make(map[string]struct{}, len(appendable))
		for _, result := range appendable {
			if err := manifest.appendResult(result.resultDir, result.resultID); err != nil {
				return err
			}
			appendedIDs[result.resultID] = struct{}{}
		}

		if len(appendedIDs) > 0 {
			metadata.Pending.ResultIDs = filterPendingResultIDs(metadata.Pending.ResultIDs, appendedIDs)
			if err := saveDownloadMetadata(runDir, *metadata); err != nil {
				return err
			}
		}

		if materializeErr != nil {
			return materializeErr
		}
	default:
		return fmt.Errorf("Unsupported download mode %q", metadata.DownloadMode)
	}

	metadata.CursorAfterID = cloneString(metadata.Pending.PageLastID)
	metadata.Pending = nil
	return saveDownloadMetadata(runDir, *metadata)
}

func orderedMaterializedPrefix(pending []downloadPipelineResultInfo, materialized []materializedPipelineResult) []materializedPipelineResult {
	materializedByID := make(map[string]materializedPipelineResult, len(materialized))
	for _, result := range materialized {
		materializedByID[result.resultID] = result
	}

	ordered := make([]materializedPipelineResult, 0, len(materialized))
	for _, result := range pending {
		materializedResult, ok := materializedByID[result.ID]
		if !ok {
			break
		}
		ordered = append(ordered, materializedResult)
	}
	return ordered
}

type materializedPipelineResult struct {
	resultID  string
	resultDir string
}

type pipelineResultMaterialization struct {
	result materializedPipelineResult
	err    error
}

func (e *downloadResultsEngine) materializePendingPipelineResults(ctx context.Context, runDir string, metadata *downloadRunMetadata, results []downloadPipelineResultInfo, workers int) ([]materializedPipelineResult, error) {
	if len(results) == 0 {
		return nil, nil
	}

	workerCount := normalizeDownloadWorkerCount(workers, len(results))
	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	jobs := make(chan downloadPipelineResultInfo)
	done := make(chan pipelineResultMaterialization, len(results))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for result := range jobs {
				materialized, err := e.materializePipelineResultArchive(ctx, runDir, metadata, result)
				done <- pipelineResultMaterialization{result: materialized, err: err}
				if err != nil {
					cancelWorkers()
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, result := range results {
			select {
			case <-workerCtx.Done():
				return
			case jobs <- result:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(done)
	}()

	materializedByID := make(map[string]materializedPipelineResult, len(results))
	var firstErr error
	for result := range done {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
				cancelWorkers()
			}
			continue
		}
		materializedByID[result.result.resultID] = result.result
	}

	materialized := make([]materializedPipelineResult, 0, len(materializedByID))
	for _, result := range results {
		if materializedResult, ok := materializedByID[result.ID]; ok {
			materialized = append(materialized, materializedResult)
		}
	}

	return materialized, firstErr
}

func (e *downloadResultsEngine) materializePipelineResultArchive(ctx context.Context, runDir string, metadata *downloadRunMetadata, result downloadPipelineResultInfo) (materializedPipelineResult, error) {
	resultDir := filepath.Join(runDir, "results", result.ID)
	archivePath, extractedDir, err := materializationPaths(resultDir, result.ArchiveURL)
	if err != nil {
		return materializedPipelineResult{}, err
	}
	if err := savePipelineResultMetadata(resultDir, result); err != nil {
		return materializedPipelineResult{}, err
	}
	if !isDownloadMaterialized(archivePath, extractedDir) {
		if err := e.materializeArchive(ctx, runDir, metadata, result.ArchiveURL, archivePath, extractedDir, map[string]any{"result_id": result.ID}); err != nil {
			return materializedPipelineResult{}, err
		}
	}
	if err := prefixPipelineResultStructureFile(resultDir, result.ID); err != nil {
		return materializedPipelineResult{}, err
	}
	return materializedPipelineResult{resultID: result.ID, resultDir: resultDir}, nil
}

func normalizeDownloadWorkerCount(workers int, jobs int) int {
	if workers < 1 {
		workers = 1
	}
	if jobs > 0 && workers > jobs {
		return jobs
	}
	return workers
}

func filterPendingResultIDs(resultIDs []string, completedIDs map[string]struct{}) []string {
	remaining := resultIDs[:0]
	for _, resultID := range resultIDs {
		if _, ok := completedIDs[resultID]; ok {
			continue
		}
		remaining = append(remaining, resultID)
	}
	return remaining
}

func (e *downloadResultsEngine) materializeArchive(ctx context.Context, runDir string, metadata *downloadRunMetadata, archiveURL string, archivePath string, extractedDir string, details map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(archivePath); errors.Is(err, os.ErrNotExist) {
		e.sink.info("archive_download", fmt.Sprintf("Downloading archive to %s", archivePath), runDir, metadata, mergeDownloadEventDetails(details, map[string]any{"archive_path": archivePath}))
		if err := e.downloadArchive(ctx, archiveURL, archivePath); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if info, err := os.Stat(extractedDir); err == nil && info.IsDir() {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	e.sink.info("archive_extract", fmt.Sprintf("Extracting archive to %s", extractedDir), runDir, metadata, mergeDownloadEventDetails(details, map[string]any{"archive_path": archivePath, "extracted_dir": extractedDir}))
	return extractArchive(archivePath, extractedDir)
}

func (e *downloadResultsEngine) downloadArchive(ctx context.Context, archiveURL string, destination string) error {
	return e.downloadArtifact(ctx, archiveURL, destination)
}

func (e *downloadResultsEngine) downloadArtifact(ctx context.Context, artifactURL string, destination string) error {
	partPath := destination + ".part"
	if err := os.Remove(partPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	file, err := os.Create(partPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var response *http.Response
	err = e.client.Get(
		ctx,
		artifactURL,
		nil,
		nil,
		option.WithResponseInto(&response),
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			r.Header.Del("Authorization")
			r.Header.Del("x-api-key")
			return next(r)
		}),
	)
	if err != nil {
		return err
	}
	if response == nil {
		return errors.New("artifact download returned no response")
	}
	defer response.Body.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(partPath, destination)
}

func extractArchive(archivePath string, destination string) error {
	if err := os.RemoveAll(destination); err != nil {
		return err
	}

	baseDir := filepath.Dir(destination)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(baseDir, filepath.Base(destination)+".tmp-")
	if err != nil {
		return err
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	switch suffix := archiveSuffixFromPath(archivePath); suffix {
	case ".tar.gz", ".tgz":
		err = extractTarGz(archivePath, tempDir)
	case ".tar":
		err = extractTar(archivePath, tempDir)
	case ".zip":
		err = extractZipArchive(archivePath, tempDir)
	default:
		err = fmt.Errorf("Unsupported archive format for %s", archivePath)
	}
	if err != nil {
		cleanup()
		return err
	}

	if err := os.RemoveAll(destination); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tempDir, destination); err != nil {
		cleanup()
		return err
	}
	return nil
}

func extractTarGz(archivePath string, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	return extractTarReader(tar.NewReader(reader), destination)
}

func extractTar(archivePath string, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return extractTarReader(tar.NewReader(file), destination)
}

func extractTarReader(reader *tar.Reader, destination string) error {
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		targetPath, err := joinArchivePath(destination, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := writeArchiveFile(targetPath, os.FileMode(header.Mode), reader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unsupported archive entry type %d in %s", header.Typeflag, header.Name)
		}
	}
}

func extractZipArchive(archivePath string, destination string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := joinArchivePath(destination, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		if err := writeArchiveFile(targetPath, file.Mode(), rc); err != nil {
			rc.Close()
			return err
		}
		if err := rc.Close(); err != nil {
			return err
		}
	}

	return nil
}

func writeArchiveFile(targetPath string, mode os.FileMode, reader io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, reader); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func joinArchivePath(destination string, entryName string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(entryName))
	if cleaned == "." {
		return destination, nil
	}
	target := filepath.Join(destination, cleaned)
	rel, err := filepath.Rel(destination, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("Archive entry escapes destination: %s", entryName)
	}
	return target, nil
}

func materializationPaths(baseDir string, archiveURL string) (string, string, error) {
	suffix, err := archiveSuffixFromURL(archiveURL)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(baseDir, "archive"+suffix), filepath.Join(baseDir, "files"), nil
}

func archiveSuffixFromURL(rawURL string) (string, error) {
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return "", err
	}
	name := path.Base(parsed.Path)
	suffix := archiveSuffixFromPath(name)
	if suffix == "" {
		return "", fmt.Errorf("Archive URL does not include a supported file suffix: %s", rawURL)
	}
	return suffix, nil
}

func archiveSuffixFromPath(rawPath string) string {
	lower := strings.ToLower(rawPath)
	for _, suffix := range downloadResultsSupportedArchiveTypes {
		if strings.HasSuffix(lower, suffix) {
			return suffix
		}
	}
	return ""
}

func isDownloadMaterialized(archivePath string, extractedDir string) bool {
	if _, err := os.Stat(archivePath); err != nil {
		return false
	}
	info, err := os.Stat(extractedDir)
	return err == nil && info.IsDir()
}

func newDownloadPipelineResultInfo(id string, archiveURL string, rawJSON string) downloadPipelineResultInfo {
	return downloadPipelineResultInfo{
		ID:         id,
		ArchiveURL: archiveURL,
		Metadata:   sanitizedDownloadJSONMap(rawJSON, id, "artifacts"),
	}
}

func savePipelineResultMetadata(resultDir string, result downloadPipelineResultInfo) error {
	metadata := cloneDownloadJSONMap(result.Metadata)
	if result.ID != "" {
		if _, ok := metadata["id"]; !ok {
			metadata["id"] = result.ID
		}
	}
	return writeDownloadJSONFile(filepath.Join(resultDir, downloadResultsResultMetadataName), metadata)
}

type pipelineResultManifestAppender struct {
	runDir   string
	runType  downloadRunType
	loaded   bool
	knownIDs map[string]struct{}
	// Column state for the small-molecule summary.csv fast path. Headers on
	// disk are cached so each new result can be appended as a single row
	// without rewriting the whole file. A row introducing a new column
	// triggers a fallback full regenerate and refreshes the cache.
	csvHeaderOrder []string
	csvHeaderSet   map[string]struct{}
}

func newPipelineResultManifestAppender(runDir string, runType downloadRunType) *pipelineResultManifestAppender {
	return &pipelineResultManifestAppender{runDir: runDir, runType: runType}
}

func (m *pipelineResultManifestAppender) appendResult(resultDir string, fallbackID string) error {
	if err := m.load(); err != nil {
		return err
	}
	if fallbackID != "" {
		if _, ok := m.knownIDs[fallbackID]; ok {
			return nil
		}
	}

	entry, err := pipelineResultManifestEntry(m.runDir, resultDir, fallbackID)
	if err != nil {
		return err
	}
	return m.appendEntry(entry)
}

func (m *pipelineResultManifestAppender) appendMetadata(metadata map[string]any, fallbackID string) error {
	if err := m.load(); err != nil {
		return err
	}
	if fallbackID != "" {
		if _, ok := m.knownIDs[fallbackID]; ok {
			return nil
		}
	}
	entry := cloneDownloadJSONMap(metadata)
	if _, ok := entry["id"]; !ok {
		entry["id"] = fallbackID
	}
	return m.appendEntry(entry)
}

func (m *pipelineResultManifestAppender) appendEntry(entry map[string]any) error {
	id := fmt.Sprint(entry["id"])
	if _, ok := m.knownIDs[id]; ok {
		return nil
	}

	indexPath := filepath.Join(m.runDir, "results", downloadResultsResultIndexName)
	lockPath := filepath.Join(m.runDir, "results", downloadResultsResultIndexLock)
	return withDownloadFileLock(lockPath, func() error {
		ids, err := readPipelineResultManifestIDs(m.runDir)
		if err != nil {
			if rebuildErr := rebuildPipelineResultManifest(m.runDir, m.runType); rebuildErr != nil {
				return fmt.Errorf("read result manifest IDs: %w; rebuild failed: %w", err, rebuildErr)
			}
			ids, err = readPipelineResultManifestIDs(m.runDir)
			if err != nil {
				return err
			}
		}
		m.knownIDs = ids
		if _, ok := m.knownIDs[id]; ok {
			return nil
		}
		if err := appendDownloadJSONLFile(indexPath, entry); err != nil {
			return err
		}
		m.knownIDs[id] = struct{}{}
		if !isSmallMoleculePipelineRunType(m.runType) {
			return nil
		}
		return m.updateSmallMoleculeSummary(entry)
	})
}

// updateSmallMoleculeSummary keeps results/summary.csv in step with the
// JSONL. If the new row's columns are already in the cached header, we
// append a single record (O(1)). A new column forces a fresh regenerate
// and refreshes the cache.
func (m *pipelineResultManifestAppender) updateSmallMoleculeSummary(entry map[string]any) error {
	flat := flattenSmallMoleculeSummaryRow(entry)
	csvPath := filepath.Join(m.runDir, "results", downloadResultsSummaryCSVName)

	if m.csvHeaderSet != nil && summaryRowFitsHeaders(flat, m.csvHeaderSet) {
		return appendSmallMoleculeSummaryRow(csvPath, m.csvHeaderOrder, flat)
	}
	return m.regenerateSmallMoleculeSummary()
}

func (m *pipelineResultManifestAppender) regenerateSmallMoleculeSummary() error {
	if err := generateSmallMoleculeSummary(m.runDir); err != nil {
		return err
	}
	return m.refreshSmallMoleculeSummaryHeaders()
}

func (m *pipelineResultManifestAppender) refreshSmallMoleculeSummaryHeaders() error {
	csvPath := filepath.Join(m.runDir, "results", downloadResultsSummaryCSVName)
	headers, err := readSmallMoleculeSummaryHeaders(csvPath)
	if err != nil {
		return err
	}
	m.csvHeaderOrder = headers
	m.csvHeaderSet = make(map[string]struct{}, len(headers))
	for _, h := range headers {
		m.csvHeaderSet[h] = struct{}{}
	}
	return nil
}

func summaryRowFitsHeaders(flat map[string]string, headerSet map[string]struct{}) bool {
	for key := range flat {
		if _, ok := headerSet[key]; !ok {
			return false
		}
	}
	return true
}

func (m *pipelineResultManifestAppender) load() error {
	if m.loaded {
		return nil
	}
	ids, err := readPipelineResultManifestIDs(m.runDir)
	if err != nil {
		if rebuildErr := rebuildPipelineResultManifest(m.runDir, m.runType); rebuildErr != nil {
			return fmt.Errorf("read result manifest IDs: %w; rebuild failed: %w", err, rebuildErr)
		}
		ids, err = readPipelineResultManifestIDs(m.runDir)
		if err != nil {
			return err
		}
	}
	m.knownIDs = ids

	// Reconcile any drift left by a prior crashed run (e.g. JSONL ahead of
	// CSV) with one full regenerate, then cache the header row so the next
	// appendEntry can take the fast path.
	if isSmallMoleculePipelineRunType(m.runType) {
		if err := m.regenerateSmallMoleculeSummary(); err != nil {
			return err
		}
	}

	m.loaded = true
	return nil
}

func readPipelineResultManifestIDs(runDir string) (map[string]struct{}, error) {
	path := filepath.Join(runDir, "results", downloadResultsResultIndexName)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]struct{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ids := map[string]struct{}{}
	decoder := json.NewDecoder(file)
	for {
		var entry map[string]any
		if err := decoder.Decode(&entry); errors.Is(err, io.EOF) {
			return ids, nil
		} else if err != nil {
			return nil, err
		}
		if id, ok := entry["id"]; ok {
			ids[fmt.Sprint(id)] = struct{}{}
		}
	}
}

func pipelineResultManifestEntry(runDir string, resultDir string, fallbackID string) (map[string]any, error) {
	metadataPath := filepath.Join(resultDir, downloadResultsResultMetadataName)
	metadata, err := readDownloadJSONFile(metadataPath)
	if err != nil {
		return nil, err
	}
	if _, ok := metadata["id"]; !ok {
		metadata["id"] = fallbackID
	}
	paths := pipelineResultLocalPaths(runDir, resultDir)
	if len(paths) > 0 {
		metadata["paths"] = paths
	}
	return metadata, nil
}

func rebuildPipelineResultManifest(runDir string, runType downloadRunType) error {
	resultsDir := filepath.Join(runDir, "results")
	entries, err := os.ReadDir(resultsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	manifest := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		resultDir := filepath.Join(resultsDir, entry.Name())
		metadata, err := pipelineResultManifestEntry(runDir, resultDir, entry.Name())
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		manifest = append(manifest, metadata)
	}

	sort.SliceStable(manifest, func(i, j int) bool {
		return fmt.Sprint(manifest[i]["id"]) < fmt.Sprint(manifest[j]["id"])
	})

	if err := writeDownloadJSONLFile(filepath.Join(resultsDir, downloadResultsResultIndexName), manifest); err != nil {
		return err
	}
	if !isSmallMoleculePipelineRunType(runType) {
		return nil
	}
	return generateSmallMoleculeSummary(runDir)
}

func pipelineResultLocalPaths(runDir string, resultDir string) map[string]string {
	paths := map[string]string{}
	if archivePath := findPipelineResultArchive(resultDir); archivePath != "" {
		paths["archive"] = relativeDownloadPath(runDir, archivePath)
	}
	extractedDir := filepath.Join(resultDir, "files")
	if info, err := os.Stat(extractedDir); err == nil && info.IsDir() {
		paths["files"] = relativeDownloadPath(runDir, extractedDir)
		if metricsPath := findExtractedResultFile(extractedDir, []string{"result/metrics.json", "metrics.json"}, func(path string) bool {
			return filepath.Base(path) == "metrics.json"
		}); metricsPath != "" {
			paths["metrics"] = relativeDownloadPath(runDir, metricsPath)
		}
		if structurePath := findExtractedResultFile(extractedDir, preferredStructureArtifactNames(filepath.Base(resultDir)), isStructureArtifactPath); structurePath != "" {
			paths["structure"] = relativeDownloadPath(runDir, structurePath)
		}
		if paePath := findExtractedResultFile(extractedDir, []string{"result/pae.npz", "pae.npz"}, func(path string) bool {
			return filepath.Base(path) == "pae.npz"
		}); paePath != "" {
			paths["pae"] = relativeDownloadPath(runDir, paePath)
		}
	}
	return paths
}

func prefixPipelineResultStructureFile(resultDir string, resultID string) error {
	return prefixStructureArtifactFile(filepath.Join(resultDir, "files"), resultID)
}

func prefixPredictionStructureFile(runDir string, predictionID string) error {
	return prefixStructureArtifactFile(filepath.Join(runDir, "outputs", "files"), predictionID)
}

func prefixStructureArtifactFile(extractedDir string, artifactID string) error {
	if info, err := os.Stat(extractedDir); err != nil || !info.IsDir() {
		return err
	}

	structurePath := findExtractedResultFile(extractedDir, preferredStructureArtifactNames(artifactID), isStructureArtifactPath)
	if structurePath == "" {
		return nil
	}

	prefixedName := prefixedStructureArtifactName(artifactID, structurePath)
	if filepath.Base(structurePath) == prefixedName {
		return nil
	}

	prefixedPath := filepath.Join(filepath.Dir(structurePath), prefixedName)
	if _, err := os.Stat(prefixedPath); err == nil {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(structurePath, prefixedPath)
}

func preferredStructureArtifactNames(resultID string) []string {
	return []string{
		path.Join("result", prefixedStructureArtifactNameWithExt(resultID, ".cif")),
		prefixedStructureArtifactNameWithExt(resultID, ".cif"),
		path.Join("result", prefixedStructureArtifactNameWithExt(resultID, ".pdb")),
		prefixedStructureArtifactNameWithExt(resultID, ".pdb"),
		"result/predicted_structure.cif",
		"predicted_structure.cif",
	}
}

func prefixedStructureArtifactName(resultID string, structurePath string) string {
	extension := strings.ToLower(filepath.Ext(structurePath))
	if extension != ".pdb" {
		extension = ".cif"
	}
	return prefixedStructureArtifactNameWithExt(resultID, extension)
}

func prefixedStructureArtifactNameWithExt(resultID string, extension string) string {
	return resultID + "_predicted" + extension
}

func findPipelineResultArchive(resultDir string) string {
	for _, suffix := range downloadResultsSupportedArchiveTypes {
		candidate := filepath.Join(resultDir, "archive"+suffix)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func findExtractedResultFile(root string, preferred []string, matches func(string) bool) string {
	for _, name := range preferred {
		candidate := filepath.Join(root, filepath.FromSlash(name))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}

	var found string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || found != "" {
			return nil
		}
		if matches(path) {
			found = path
		}
		return nil
	})
	return found
}

func isStructureArtifactPath(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == "predicted_structure.cif" || strings.HasSuffix(name, ".cif") || strings.HasSuffix(name, ".pdb")
}

func relativeDownloadPath(root string, target string) string {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return filepath.ToSlash(target)
	}
	return filepath.ToSlash(rel)
}

func sanitizedDownloadJSONMap(rawJSON string, fallbackID string, omitTopLevelKeys ...string) map[string]any {
	var payload map[string]any
	if strings.TrimSpace(rawJSON) != "" {
		if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
			payload = nil
		}
	}
	if payload == nil {
		payload = map[string]any{}
	}
	for _, key := range omitTopLevelKeys {
		delete(payload, key)
	}
	scrubPresignedURLFields(payload)
	if fallbackID != "" {
		if _, ok := payload["id"]; !ok {
			payload["id"] = fallbackID
		}
	}
	return payload
}

func scrubPresignedURLFields(value any) {
	switch typed := value.(type) {
	case map[string]any:
		delete(typed, "url")
		delete(typed, "url_expires_at")
		for _, child := range typed {
			scrubPresignedURLFields(child)
		}
	case []any:
		for _, child := range typed {
			scrubPresignedURLFields(child)
		}
	}
}

func cloneDownloadJSONMap(value map[string]any) map[string]any {
	cloned := make(map[string]any, len(value))
	for key, child := range value {
		cloned[key] = child
	}
	return cloned
}

func readDownloadJSONFile(path string) (map[string]any, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func writeDownloadJSONFile(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return writeDownloadFileAtomically(path, payload)
}

func writeDownloadJSONLFile(path string, values []map[string]any) error {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	encoder.SetEscapeHTML(false)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			return err
		}
	}
	return writeDownloadFileAtomically(path, []byte(builder.String()))
}

func appendDownloadJSONLFile(path string, value map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	return appendDownloadJSONL(file, value)
}

func withDownloadFileLock(path string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	deadline := time.Now().Add(downloadResultsManifestLockWait)
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
			closeErr := file.Close()
			defer os.Remove(path)
			if closeErr != nil {
				return closeErr
			}
			return fn()
		}
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Timed out waiting for download manifest lock: %s", path)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func appendDownloadJSONL(writer io.WriteCloser, value map[string]any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}
	return writer.Close()
}

func writeDownloadFileAtomically(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func updateRunMetadata(metadata *downloadRunMetadata, status string, workspaceID *string, startedAt *string, completedAt *string, stoppedAt *string, latestResultID *string, errorCode *string) {
	if workspaceID != nil {
		metadata.Remote.WorkspaceID = cloneString(workspaceID)
	}
	metadata.Remote.Status = cloneString(&status)
	metadata.Remote.StartedAt = cloneString(startedAt)
	metadata.Remote.CompletedAt = cloneString(completedAt)
	metadata.Remote.StoppedAt = cloneString(stoppedAt)
	metadata.Remote.LatestResultID = cloneString(latestResultID)
	metadata.Remote.ErrorCode = cloneString(errorCode)
	if metadata.Remote.RunID != nil {
		metadata.RequestFingerprint = fingerprintDownloadRun(metadata.RunType, *metadata.Remote.RunID, metadata.Remote.WorkspaceID)
	}
}

func failureForMetadata(metadata downloadRunMetadata) error {
	runID := metadata.Name
	if metadata.Remote.RunID != nil && *metadata.Remote.RunID != "" {
		runID = *metadata.Remote.RunID
	}
	if metadata.Remote.ErrorCode != nil && *metadata.Remote.ErrorCode != "" {
		return fmt.Errorf("Run %s failed with error code %s", runID, *metadata.Remote.ErrorCode)
	}
	return fmt.Errorf("Run %s failed", runID)
}

func pipelineRunningSummary(latestResultID *string) string {
	if latestResultID == nil || *latestResultID == "" {
		return "Pipeline still running..."
	}
	return fmt.Sprintf("Pipeline still running... latest result: %s", *latestResultID)
}

func (s *downloadResultsSink) info(event string, message string, runDir string, metadata *downloadRunMetadata, details map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.progressFormat == downloadProgressFormatJSONL {
		s.writeJSONL(event, message, runDir, metadata, details)
		return
	}
	if !s.verbose {
		return
	}
	_, _ = fmt.Fprintln(s.writer, message)
}

func (s *downloadResultsSink) maybeRunningSummary(event string, message string, runDir string, metadata *downloadRunMetadata, details map[string]any) {
	if s.progressFormat != downloadProgressFormatJSONL && !s.verbose {
		return
	}
	now := time.Now()
	s.mu.Lock()
	if !s.lastRunningSummaryAt.IsZero() && now.Sub(s.lastRunningSummaryAt) < downloadResultsSummaryIntervalSec {
		s.mu.Unlock()
		return
	}
	s.lastRunningSummaryAt = now
	s.mu.Unlock()
	s.info(event, message, runDir, metadata, details)
}

func (s *downloadResultsSink) writeJSONL(event string, message string, runDir string, metadata *downloadRunMetadata, details map[string]any) {
	payload := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"event": event,
	}
	if message != "" {
		payload["message"] = message
	}
	if metadata != nil {
		mergeDownloadEventFields(payload, buildDownloadStatusResponse(runDir, *metadata))
	} else if runDir != "" {
		payload["run_dir"] = runDir
	}
	for key, value := range details {
		if value == nil {
			continue
		}
		payload[key] = value
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintln(s.writer, string(body))
}

func mergeDownloadEventFields(target map[string]any, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		return
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		return
	}
	for key, value := range fields {
		target[key] = value
	}
}

func mergeDownloadEventDetails(base map[string]any, extra map[string]any) map[string]any {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}
	merged := make(map[string]any, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func (e *downloadResultsEngine) getPrediction(ctx context.Context, runID string, workspaceID *string) (downloadPredictionRunInfo, error) {
	params := boltzapi.PredictionStructureAndBindingGetParams{}
	setOptionalStringOpt(&params.WorkspaceID, workspaceID)
	response, err := e.client.Predictions.StructureAndBinding.Get(ctx, runID, params)
	if err != nil {
		return downloadPredictionRunInfo{}, err
	}
	return downloadPredictionRunInfo{
		Status:      string(response.Status),
		WorkspaceID: normalizedStringPointer(response.WorkspaceID),
		StartedAt:   normalizedTimePointer(response.StartedAt),
		CompletedAt: normalizedTimePointer(response.CompletedAt),
		ErrorCode:   normalizedStringPointer(response.Error.Code),
		ArchiveURL:  normalizedStringPointer(response.Output.Archive.URL),
		Run:         sanitizedDownloadJSONMap(response.RawJSON(), response.ID),
	}, nil
}

func (e *downloadResultsEngine) getAdmePrediction(ctx context.Context, runID string, workspaceID *string) (downloadPredictionRunInfo, error) {
	params := boltzapi.PredictionAdmeGetParams{}
	setOptionalStringOpt(&params.WorkspaceID, workspaceID)
	response, err := e.client.Predictions.Adme.Get(ctx, runID, params)
	if err != nil {
		return downloadPredictionRunInfo{}, err
	}
	return downloadPredictionRunInfo{
		Status:      string(response.Status),
		WorkspaceID: normalizedStringPointer(response.WorkspaceID),
		StartedAt:   normalizedTimePointer(response.StartedAt),
		CompletedAt: normalizedTimePointer(response.CompletedAt),
		ErrorCode:   normalizedStringPointer(response.Error.Code),
		Run:         sanitizedDownloadJSONMap(response.RawJSON(), response.ID),
	}, nil
}

func (e *downloadResultsEngine) getPipeline(ctx context.Context, runType downloadRunType, runID string, workspaceID *string) (downloadPipelineRunInfo, error) {
	adapter, err := pipelineAdapterFor(runType)
	if err != nil {
		return downloadPipelineRunInfo{}, err
	}
	return adapter.get(ctx, e.client, runID, workspaceID)
}

func (e *downloadResultsEngine) listPipelineResults(ctx context.Context, runType downloadRunType, runID string, workspaceID *string, afterID *string) (downloadPipelinePage, error) {
	adapter, err := pipelineAdapterFor(runType)
	if err != nil {
		return downloadPipelinePage{}, err
	}
	return adapter.listResults(ctx, e.client, runID, workspaceID, afterID)
}

func newDownloadPipelineRunInfo(status string, runID string, workspaceID string, startedAt time.Time, completedAt time.Time, stoppedAt time.Time, latestResultID string, errorCode string, rawJSON string) downloadPipelineRunInfo {
	return downloadPipelineRunInfo{
		Status:         status,
		WorkspaceID:    normalizedStringPointer(workspaceID),
		StartedAt:      normalizedTimePointer(startedAt),
		CompletedAt:    normalizedTimePointer(completedAt),
		StoppedAt:      normalizedTimePointer(stoppedAt),
		LatestResultID: normalizedStringPointer(latestResultID),
		ErrorCode:      normalizedStringPointer(errorCode),
		Run:            sanitizedDownloadJSONMap(rawJSON, runID),
	}
}

func normalizePipelinePage[T any](page *pagination.CursorPage[T], normalize func(T) downloadPipelineResultInfo) downloadPipelinePage {
	results := make([]downloadPipelineResultInfo, 0, len(page.Data))
	for _, result := range page.Data {
		results = append(results, normalize(result))
	}
	return downloadPipelinePage{
		Results: results,
		LastID:  trimOptionalString(page.LastID),
	}
}

func pipelineAdapterFor(runType downloadRunType) (downloadPipelineAdapter, error) {
	adapter, ok := downloadPipelineAdapters[runType]
	if !ok {
		return downloadPipelineAdapter{}, fmt.Errorf("Unsupported run type %q", runType)
	}
	return adapter, nil
}

func setOptionalStringOpt(target *param.Opt[string], value *string) {
	if value == nil {
		return
	}
	*target = param.NewOpt(*value)
}

func normalizedStringPointer(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizedTimePointer(value time.Time) *string {
	if value.IsZero() {
		return nil
	}
	formatted := value.Format(time.RFC3339Nano)
	return &formatted
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func trimOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func sleepWithContext(ctx context.Context, seconds float64) error {
	if seconds <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(time.Duration(seconds * float64(time.Second)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
