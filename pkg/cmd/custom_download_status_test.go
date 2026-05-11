// Custom CLI extension code. Not generated.
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadStatusCommandValidation(t *testing.T) {
	setDownloadResultsTestEnv(t)
	t.Chdir(t.TempDir())

	testCases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing selector",
			args:    []string{"download-status"},
			wantErr: "--name or --run-dir is required",
		},
		{
			name:    "name and run-dir are mutually exclusive",
			args:    []string{"download-status", "--name", "foo", "--run-dir", "bar"},
			wantErr: "--name and --run-dir are mutually exclusive",
		},
		{
			name:    "run-dir rejects explicit root-dir",
			args:    []string{"download-status", "--run-dir", "bar", "--root-dir", "runs"},
			wantErr: "--root-dir cannot be used with --run-dir",
		},
		{
			name:    "extra positional arguments",
			args:    []string{"download-status", "foo"},
			wantErr: "Use --name or --run-dir to select a local run",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, stderr, err := runDownloadStatusCLI(t, testCase.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.wantErr)
			assert.Empty(t, stdout)
			assert.Empty(t, stderr)
		})
	}
}

func TestDownloadStatusReportsStructuredMetadataSnapshot(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "status-run")
	require.NoError(t, os.MkdirAll(runDir, 0o755))

	metadata := newDownloadRunMetadata("status-run", downloadRunTypeProteinDesign, "prot_des_123", nil, downloadModeEverything)
	status := "running"
	startedAt := "2026-04-22T10:00:00Z"
	latestResultID := "res_7"
	cursorAfterID := "res_5"
	pageLastID := "res_7"
	metadata.Remote.Status = &status
	metadata.Remote.StartedAt = &startedAt
	metadata.Remote.LatestResultID = &latestResultID
	metadata.CursorAfterID = &cursorAfterID
	metadata.Pending = &downloadPendingState{
		Kind:       downloadPendingKindResultPage,
		AfterID:    &cursorAfterID,
		PageLastID: &pageLastID,
		ResultIDs:  []string{"res_6", "res_7"},
	}
	require.NoError(t, saveDownloadMetadata(runDir, metadata))

	stdout, stderr, err := runDownloadStatusCLI(t, "--format", "json", "download-status", "--name", "status-run")
	require.NoError(t, err)
	assert.Empty(t, stderr)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &response))
	assert.Equal(t, runDir, response["run_dir"])
	assert.Equal(t, "status-run", response["name"])
	assert.Equal(t, string(downloadRunTypeProteinDesign), response["run_type"])
	assert.Equal(t, downloadModeEverything, response["download_mode"])
	assert.Equal(t, "prot_des_123", response["run_id"])
	assert.Equal(t, "running", response["status"])
	assert.Equal(t, "materializing_results", response["phase"])
	assert.Equal(t, false, response["ready"])
	assert.Equal(t, float64(2), response["pending_count"])
	assert.Equal(t, "res_7", response["latest_result_id"])

	pendingResultIDs, ok := response["pending_result_ids"].([]any)
	require.True(t, ok)
	assert.Len(t, pendingResultIDs, 2)
}

func runDownloadStatusCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	var stderr bytes.Buffer
	root := newDownloadResultsTestRoot(w, &stderr)
	runErr := root.Run(context.Background(), append([]string{"boltz-api"}, args...))
	require.NoError(t, w.Close())

	stdout, readErr := io.ReadAll(r)
	require.NoError(t, readErr)
	return string(stdout), stderr.String(), runErr
}
