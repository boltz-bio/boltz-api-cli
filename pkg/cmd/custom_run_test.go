// Custom CLI extension code. Not generated.
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestApplyCustomizationsAddsRunCommands(t *testing.T) {
	ApplyCustomizations(Command)

	testCases := []struct {
		path               []string
		pipeline           bool
		innerFlags         []string
		inputUsageContains []string
	}{
		{
			path:       []string{"predictions:structure-and-binding"},
			innerFlags: []string{"input.entities", "input.num-samples"},
			inputUsageContains: []string{
				"Structure and binding prediction input",
				"inline JSON/YAML",
				"@yaml://",
				"entities",
				"templates",
				"--model",
				"--workspace-id",
			},
		},
		{
			path:       []string{"predictions:adme"},
			innerFlags: []string{"input.molecules"},
			inputUsageContains: []string{
				"ADME prediction input",
				"inline JSON/YAML",
				"@yaml://",
				"molecules",
				"SMILES",
				"--model",
				"--workspace-id",
			},
		},
		{
			path:       []string{"protein:design"},
			pipeline:   true,
			innerFlags: []string{"input.binder-specification", "input.num-proteins", "input.target"},
		},
		{
			path:       []string{"protein:library-screen"},
			pipeline:   true,
			innerFlags: []string{"input.proteins", "input.target"},
		},
		{
			path:       []string{"small-molecule:design"},
			pipeline:   true,
			innerFlags: []string{"input.num-molecules", "input.target", "input.chemical-space"},
		},
		{
			path:       []string{"small-molecule:library-screen"},
			pipeline:   true,
			innerFlags: []string{"input.molecules", "input.target"},
		},
	}

	for _, testCase := range testCases {
		t.Run(strings.Join(testCase.path, " "), func(t *testing.T) {
			parent := mustFindCommand(t, Command, testCase.path...)
			run := mustFindCommand(t, Command, append(testCase.path, "run")...)

			requireRunCommandFollowsStart(t, parent)
			mustFindFlag(t, run, "name")
			mustFindFlag(t, run, "run-dir")
			mustFindFlag(t, run, "root-dir")
			mustFindFlag(t, run, "poll-interval-seconds")
			mustFindFlag(t, run, "progress-format")
			mustFindFlag(t, run, "verbose")
			mustFindFlag(t, run, "workspace-id")

			if testCase.pipeline {
				mustFindFlag(t, run, "download-mode")
				mustFindFlag(t, run, "workers")
			} else {
				require.Nil(t, findFlag(run, "download-mode"))
				require.Nil(t, findFlag(run, "workers"))
			}

			for _, name := range testCase.innerFlags {
				mustFindFlag(t, run, name)
			}
			if len(testCase.inputUsageContains) > 0 {
				inputUsage := usageForFlag(t, mustFindFlag(t, run, "input"))
				for _, expected := range testCase.inputUsageContains {
					require.Contains(t, inputUsage, expected)
				}
			}
			require.NoError(t, requestflag.CheckInnerFlags(*run))
		})
	}
}

func TestAdmeRunStartsWaitsAndWritesInlineOutput(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	const runID = "adme_pred_123"
	output := map[string]any{
		"molecules": []map[string]any{{
			"id":          "mol_1",
			"external_id": "aspirin",
			"smiles":      "CC(=O)OC1=CC=CC=C1C(=O)O",
			"status":      "succeeded",
			"adme": map[string]any{
				"lipophilicity": 0.1,
				"permeability":  0.2,
				"solubility":    "high-confidence",
			},
			"error": nil,
		}},
	}

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/adme":
			require.Equal(t, http.MethodPost, r.Method)
			body := decodeRequestJSON(t, r)
			require.Equal(t, "adme-v1", body["model"])
			require.Equal(t, "ws_request", body["workspace_id"])
			input, ok := body["input"].(map[string]any)
			require.True(t, ok)
			molecules, ok := input["molecules"].([]any)
			require.True(t, ok)
			require.Len(t, molecules, 1)
			writeJSON(t, w, admePredictionResponseJSON(runID, "running", "ws_123", nil, ""))
		case "/compute/v1/predictions/adme/" + runID:
			require.Equal(t, "ws_123", r.URL.Query().Get("workspace_id"))
			writeJSON(t, w, admePredictionResponseJSON(runID, "succeeded", "ws_123", output, ""))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runRunCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"predictions:adme",
		"run",
		"--input.molecules", "[{smiles: 'CC(=O)OC1=CC=CC=C1C(=O)O', id: aspirin}]",
		"--model", "adme-v1",
		"--workspace-id", "ws_request",
		"--name", "adme-run",
		"--poll-interval-seconds", "0",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "adme-run")
	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.NoDirExists(t, filepath.Join(runDir, "outputs"))

	metadata := mustLoadDownloadMetadata(t, runDir)
	assert.Equal(t, downloadRunTypeAdme, metadata.RunType)
	assert.Equal(t, runID, derefString(metadata.Remote.RunID))
	assert.Equal(t, "succeeded", derefString(metadata.Remote.Status))

	runJSON := readDownloadResultsTestJSON(t, filepath.Join(runDir, "run.json"))
	assert.Equal(t, runID, runJSON["id"])
	runOutput, ok := runJSON["output"].(map[string]any)
	require.True(t, ok)
	molecules, ok := runOutput["molecules"].([]any)
	require.True(t, ok)
	require.Len(t, molecules, 1)
	molecule, ok := molecules[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "aspirin", molecule["external_id"])
	assert.Equal(t, "succeeded", molecule["status"])
}

func TestPipelineRunStartsWaitsAndDownloadsResults(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	const runID = "prot_des_123"
	archiveBytes := makeTarGzArchive(t, map[string]string{"result.txt": "protein design"})

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/protein/design":
			require.Equal(t, http.MethodPost, r.Method)
			body := decodeRequestJSON(t, r)
			require.EqualValues(t, 10, body["num_proteins"])
			require.Equal(t, "ws_request", body["workspace_id"])
			require.NotNil(t, body["binder_specification"])
			require.NotNil(t, body["target"])
			writeJSON(t, w, pipelineGetResponseJSON(runID, "running", "ws_123", "res_1", ""))
		case "/compute/v1/protein/design/" + runID:
			require.Equal(t, "ws_123", r.URL.Query().Get("workspace_id"))
			writeJSON(t, w, pipelineGetResponseJSON(runID, "succeeded", "ws_123", "res_1", ""))
		case "/compute/v1/protein/design/" + runID + "/results":
			require.Equal(t, "ws_123", r.URL.Query().Get("workspace_id"))
			require.Equal(t, fmt.Sprint(downloadResultsPageLimit), r.URL.Query().Get("limit"))
			if r.URL.Query().Get("after_id") == "res_1" {
				writeJSON(t, w, emptyPipelinePageJSON())
				return
			}
			writeJSON(t, w, pipelineResultsPageJSON(server.URL, runID, "res_1"))
		case "/files/" + runID + "/res_1.tar.gz":
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runRunCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"protein:design",
		"run",
		"--input", "{binder_specification: {type: no_template, modality: peptide}, num_proteins: 10, target: {type: structure_template, structure: {type: url, url: 'https://example.com/target.cif'}}}",
		"--workspace-id", "ws_request",
		"--name", "protein-run",
		"--poll-interval-seconds", "0",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "protein-run")
	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.FileExists(t, filepath.Join(runDir, "run.json"))
	assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "archive.tar.gz"))
	assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "files", "result.txt"))

	metadata := mustLoadDownloadMetadata(t, runDir)
	assert.Equal(t, downloadRunTypeProteinDesign, metadata.RunType)
	assert.Equal(t, downloadModeEverything, metadata.DownloadMode)
	assert.Equal(t, "succeeded", derefString(metadata.Remote.Status))

	manifest := readDownloadResultsTestJSONL(t, filepath.Join(runDir, "results", "index.jsonl"))
	require.Len(t, manifest, 1)
	assert.Equal(t, "res_1", manifest[0]["id"])
}

func TestRunRejectsExistingMetadataBeforeStartingRemoteRun(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "existing")
	require.NoError(t, ensureDownloadDirectoryReady(runDir))
	require.NoError(t, saveDownloadMetadata(runDir, newDownloadRunMetadata("existing", downloadRunTypeAdme, "adme_pred_old", nil, downloadModeEverything)))

	startRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startRequests++
		http.Error(w, "should not start", http.StatusInternalServerError)
	}))
	defer server.Close()

	stdout, stderr, err := runRunCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"predictions:adme",
		"run",
		"--input.molecules", "[{smiles: CCO}]",
		"--model", "adme-v1",
		"--name", "existing",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already contains .boltz-run.json")
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Equal(t, 0, startRequests)
}

func TestRunValidatesDeterministicRootDirBeforeStartingRemoteRun(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	badRoot := filepath.Join(cwd, "not-a-dir")
	require.NoError(t, os.WriteFile(badRoot, []byte("file"), 0o644))

	startRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startRequests++
		http.Error(w, "should not start", http.StatusInternalServerError)
	}))
	defer server.Close()

	stdout, stderr, err := runRunCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"predictions:adme",
		"run",
		"--input.molecules", "[{smiles: CCO}]",
		"--model", "adme-v1",
		"--root-dir", badRoot,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Run root path is not a directory")
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
	assert.Equal(t, 0, startRequests)
}

func runRunCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := newRunTestRoot(&stdout, &stderr)
	err := root.Run(context.Background(), append([]string{"boltz-api"}, args...))
	return stdout.String(), stderr.String(), err
}

func newRunTestRoot(stdout io.Writer, stderr io.Writer) *cli.Command {
	root := &cli.Command{
		Name:      "boltz-api",
		Writer:    stdout,
		ErrWriter: stderr,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "debug"},
			&cli.StringFlag{Name: "base-url"},
			&cli.StringFlag{Name: "format", Value: "auto"},
			&cli.StringFlag{Name: "transform"},
			&cli.BoolFlag{Name: "raw-output"},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_API_KEY"),
			},
		},
		Commands: []*cli.Command{
			{
				Name: "predictions:adme",
				Commands: []*cli.Command{
					&predictionsAdmeStart,
				},
			},
			{
				Name: "protein:design",
				Commands: []*cli.Command{
					&proteinDesignStart,
				},
			},
		},
	}
	ApplyCustomizations(root)
	return root
}

func decodeRequestJSON(t *testing.T, r *http.Request) map[string]any {
	t.Helper()

	var payload map[string]any
	require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
	require.NotNil(t, payload)
	return payload
}

func admePredictionResponseJSON(id string, status string, workspaceID string, output any, errorCode string) map[string]any {
	payload := map[string]any{
		"id":           id,
		"status":       status,
		"workspace_id": workspaceID,
		"started_at":   time.Now().UTC().Format(time.RFC3339),
		"completed_at": time.Now().UTC().Format(time.RFC3339),
		"input": map[string]any{
			"molecules": []map[string]any{{
				"id":     "aspirin",
				"smiles": "CC(=O)OC1=CC=CC=C1C(=O)O",
			}},
		},
		"model": "adme-v1",
	}
	if output != nil {
		payload["output"] = output
	}
	if errorCode != "" {
		payload["error"] = map[string]any{"code": errorCode}
	}
	return payload
}

func requireRunCommandFollowsStart(t *testing.T, parent *cli.Command) {
	t.Helper()

	startIndex := -1
	runIndex := -1
	for i, child := range parent.Commands {
		if child == nil {
			continue
		}
		switch child.Name {
		case "start":
			startIndex = i
		case "run":
			runIndex = i
		}
	}

	require.NotEqual(t, -1, startIndex)
	require.Equal(t, startIndex+1, runIndex)
}
