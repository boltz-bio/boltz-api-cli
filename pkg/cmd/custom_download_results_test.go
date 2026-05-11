// Custom CLI extension code. Not generated.
package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestDownloadResultsCommandValidation(t *testing.T) {
	setDownloadResultsTestEnv(t)
	t.Chdir(t.TempDir())

	testCases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "positional id requires flag",
			args:    []string{"download-results", "pred_123"},
			wantErr: "Use --id for the run ID",
		},
		{
			name:    "extra positional arguments",
			args:    []string{"download-results", "pred_123", "unexpected"},
			wantErr: "Unexpected extra arguments",
		},
		{
			name:    "name and run-dir are mutually exclusive",
			args:    []string{"download-results", "--id", "pred_123", "--name", "foo", "--run-dir", "bar"},
			wantErr: "--name and --run-dir are mutually exclusive",
		},
		{
			name:    "run-dir rejects explicit root-dir",
			args:    []string{"download-results", "--id", "pred_123", "--run-dir", "bar", "--root-dir", "runs"},
			wantErr: "--root-dir cannot be used with --run-dir",
		},
		{
			name:    "missing id without local metadata",
			args:    []string{"download-results", "--name", "missing"},
			wantErr: "No local metadata exists",
		},
		{
			name:    "reject pipeline result ids",
			args:    []string{"download-results", "--id", "pres_123"},
			wantErr: "download-results only supports prediction and pipeline run IDs",
		},
		{
			name:    "reject pipeline sample ids",
			args:    []string{"download-results", "--id", "psmp_123"},
			wantErr: "download-results only supports prediction and pipeline run IDs",
		},
		{
			name:    "reject invalid progress format",
			args:    []string{"download-results", "--id", "pred_123", "--progress-format", "csv"},
			wantErr: "--progress-format must be one of",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, stderr, err := runDownloadResultsCLI(t, testCase.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.wantErr)
			assert.Empty(t, stdout)
			assert.Empty(t, stderr)
		})
	}
}

func TestDownloadResultsMetadataRoundTripPreservesExtraFields(t *testing.T) {
	runDir := t.TempDir()

	metadata := newDownloadRunMetadata("example", downloadRunTypePrediction, "pred_123", nil)
	metadata.ExtraFields = map[string]json.RawMessage{
		"unexpected": json.RawMessage(`true`),
	}

	require.NoError(t, saveDownloadMetadata(runDir, metadata))
	loaded, err := loadDownloadMetadata(runDir)
	require.NoError(t, err)
	require.Contains(t, loaded.ExtraFields, "unexpected")

	require.NoError(t, saveDownloadMetadata(runDir, loaded))

	body, err := os.ReadFile(downloadMetadataPath(runDir))
	require.NoError(t, err)
	assert.Contains(t, string(body), `"unexpected": true`)
}

func TestInferDownloadRunTypeSupportsStructurePredictionPrefixes(t *testing.T) {
	testCases := []string{
		"sab_pred_123",
		"pred_123",
	}

	for _, runID := range testCases {
		t.Run(runID, func(t *testing.T) {
			runType, err := inferDownloadRunType(runID)
			require.NoError(t, err)
			assert.Equal(t, downloadRunTypePrediction, runType)
		})
	}
}

func TestPredictionDownloadResultsDefaultProgressJSONL(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var archiveRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			archiveRequests.Add(1)
			assert.Empty(t, r.Header.Get("Authorization"))
			assert.Empty(t, r.Header.Get("x-api-key"))
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "prediction-run",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "prediction-run")
	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.EqualValues(t, 1, archiveRequests.Load())
	assert.FileExists(t, filepath.Join(runDir, "outputs", "archive.tar.gz"))
	assert.FileExists(t, filepath.Join(runDir, "outputs", "files", "nested", "output.txt"))

	run := readDownloadResultsTestJSON(t, filepath.Join(runDir, "run.json"))
	assert.Equal(t, "pred_123", run["id"])
	input, ok := run["input"].(map[string]any)
	require.True(t, ok)
	entities, ok := input["entities"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, entities)
	runBody, err := os.ReadFile(filepath.Join(runDir, "run.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(runBody), server.URL)
}

func TestStructureAndBindingPredictionPrefixDownloadsResults(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runID := "sab_pred_123"
	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var archiveRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/" + runID:
			writeJSON(t, w, predictionResponseJSON(runID, "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			archiveRequests.Add(1)
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", runID,
		"--name", "sab-prediction-run",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "sab-prediction-run")
	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.EqualValues(t, 1, archiveRequests.Load())
	assert.FileExists(t, filepath.Join(runDir, "outputs", "archive.tar.gz"))
	assert.FileExists(t, filepath.Join(runDir, "outputs", "files", "nested", "output.txt"))
}

func TestPredictionDownloadResultsUsesExistingDirectoryWithoutMetadata(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "prediction-run")
	require.NoError(t, os.MkdirAll(runDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "notes.txt"), []byte("precreated"), 0o644))

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var archiveRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			archiveRequests.Add(1)
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "prediction-run",
	)
	require.NoError(t, err)

	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.EqualValues(t, 1, archiveRequests.Load())
	assert.FileExists(t, filepath.Join(runDir, "notes.txt"))
	assert.FileExists(t, downloadMetadataPath(runDir))
	assert.FileExists(t, filepath.Join(runDir, "outputs", "archive.tar.gz"))
	assert.FileExists(t, filepath.Join(runDir, "outputs", "files", "nested", "output.txt"))
}

func TestPredictionDownloadResultsSupportsSabPredPrefix(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	const runID = "sab_pred_123"

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var archiveRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/" + runID:
			writeJSON(t, w, predictionResponseJSON(runID, "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			archiveRequests.Add(1)
			assert.Empty(t, r.Header.Get("Authorization"))
			assert.Empty(t, r.Header.Get("x-api-key"))
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", runID,
		"--name", "prediction-run",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "prediction-run")
	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.EqualValues(t, 1, archiveRequests.Load())
	assert.FileExists(t, filepath.Join(runDir, "outputs", "archive.tar.gz"))
	assert.FileExists(t, filepath.Join(runDir, "outputs", "files", "nested", "output.txt"))
}

func TestPredictionResumeReextractsMissingDirectory(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var archiveRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			archiveRequests.Add(1)
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, _, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "resume-prediction",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "resume-prediction")
	extractedDir := filepath.Join(runDir, "outputs", "files")
	require.NoError(t, os.RemoveAll(extractedDir))

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--name", "resume-prediction",
	)
	require.NoError(t, err)

	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.EqualValues(t, 1, archiveRequests.Load())
	assert.DirExists(t, extractedDir)
	assert.FileExists(t, filepath.Join(extractedDir, "nested", "output.txt"))
}

func TestPredictionPartialDownloadLeavesPartAndRecovers(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var archiveRequests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			if archiveRequests.Add(1) == 1 {
				writePartialArchiveAndClose(t, w, archiveBytes[:12], len(archiveBytes))
				return
			}
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, _, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "partial-prediction",
	)
	require.Error(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "partial-prediction")
	archivePath := filepath.Join(runDir, "outputs", "archive.tar.gz")
	assert.NoFileExists(t, archivePath)
	assert.FileExists(t, archivePath+".part")

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--name", "partial-prediction",
	)
	require.NoError(t, err)

	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.EqualValues(t, 2, archiveRequests.Load())
	assert.FileExists(t, archivePath)
	assert.NoFileExists(t, archivePath+".part")
	assert.FileExists(t, filepath.Join(runDir, "outputs", "files", "nested", "output.txt"))
}

func TestPredictionSucceededWithoutArchiveURLFails(t *testing.T) {
	setDownloadResultsTestEnv(t)
	t.Chdir(t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", "", ""))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, _, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "missing-archive",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not return an archive URL")
}

func TestPipelineDownloadResultsHappyPathAllRunTypes(t *testing.T) {
	testCases := []struct {
		name        string
		runID       string
		getPath     string
		resultsPath string
	}{
		{name: "protein design", runID: "prot_des_123", getPath: "/compute/v1/protein/design/prot_des_123", resultsPath: "/compute/v1/protein/design/prot_des_123/results"},
		{name: "protein library screen", runID: "prot_scr_123", getPath: "/compute/v1/protein/library-screen/prot_scr_123", resultsPath: "/compute/v1/protein/library-screen/prot_scr_123/results"},
		{name: "small molecule design", runID: "sm_des_123", getPath: "/compute/v1/small-molecule/design/sm_des_123", resultsPath: "/compute/v1/small-molecule/design/sm_des_123/results"},
		{name: "small molecule library screen", runID: "sm_scr_123", getPath: "/compute/v1/small-molecule/library-screen/sm_scr_123", resultsPath: "/compute/v1/small-molecule/library-screen/sm_scr_123/results"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			setDownloadResultsTestEnv(t)
			cwd := t.TempDir()
			t.Chdir(cwd)

			archiveBytes := makeTarGzArchive(t, map[string]string{
				"result.txt":                     testCase.name,
				"result/metrics.json":            `{"score":0.99}`,
				"result/predicted_structure.cif": "data_test\n#\n",
				"result/pae.npz":                 "pae",
			})
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case testCase.getPath:
					writeJSON(t, w, pipelineGetResponseJSON(testCase.runID, "succeeded", "ws_123", "res_2", ""))
				case testCase.resultsPath:
					require.Equal(t, "20", r.URL.Query().Get("limit"))
					if r.URL.Query().Get("after_id") == "res_2" {
						writeJSON(t, w, emptyPipelinePageJSON())
						return
					}
					writeJSON(t, w, pipelineResultsPageJSON(server.URL, testCase.runID, "res_1", "res_2"))
				case "/files/" + testCase.runID + "/res_1.tar.gz", "/files/" + testCase.runID + "/res_2.tar.gz":
					_, _ = w.Write(archiveBytes)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			runName := strings.ReplaceAll(testCase.runID, "_", "-")
			stdout, stderr, err := runDownloadResultsCLI(
				t,
				"--base-url", server.URL,
				"--api-key", "test-key",
				"download-results",
				"--id", testCase.runID,
				"--name", runName,
			)
			require.NoError(t, err)

			runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, runName)
			assert.Equal(t, runDir+"\n", stdout)
			assert.NotEmpty(t, stderr)
			assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "archive.tar.gz"))
			assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "files", "result.txt"))
			assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "metadata.json"))
			assert.FileExists(t, filepath.Join(runDir, "results", "res_2", "archive.tar.gz"))
			assert.FileExists(t, filepath.Join(runDir, "results", "res_2", "files", "result.txt"))
			assert.FileExists(t, filepath.Join(runDir, "run.json"))

			metadata := readDownloadResultsTestJSON(t, filepath.Join(runDir, "results", "res_1", "metadata.json"))
			assert.Equal(t, "res_1", metadata["id"])
			assert.NotContains(t, metadata, "artifacts")
			assertPipelineResultMetadata(t, testCase.runID, metadata)

			metadataBody, err := os.ReadFile(filepath.Join(runDir, "results", "res_1", "metadata.json"))
			require.NoError(t, err)
			assert.NotContains(t, string(metadataBody), server.URL)

			manifest := readDownloadResultsTestJSONL(t, filepath.Join(runDir, "results", "index.jsonl"))
			require.Len(t, manifest, 2)
			assert.Equal(t, "res_1", manifest[0]["id"])
			paths, ok := manifest[0]["paths"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "results/res_1/archive.tar.gz", paths["archive"])
			assert.Equal(t, "results/res_1/files/result/metrics.json", paths["metrics"])
			assert.Equal(t, "results/res_1/files/result/predicted_structure.cif", paths["structure"])
			assert.Equal(t, "results/res_1/files/result/pae.npz", paths["pae"])
		})
	}
}

func TestPipelineDownloadResultsResumesPendingPageAfterPartialFailure(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runID := "prot_des_123"
	archiveBytes := makeTarGzArchive(t, map[string]string{"result.txt": "resume"})
	var archive2Requests atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/protein/design/" + runID:
			writeJSON(t, w, pipelineGetResponseJSON(runID, "succeeded", "ws_123", "res_2", ""))
		case "/compute/v1/protein/design/" + runID + "/results":
			require.Equal(t, "20", r.URL.Query().Get("limit"))
			if r.URL.Query().Get("after_id") == "res_2" {
				writeJSON(t, w, emptyPipelinePageJSON())
				return
			}
			writeJSON(t, w, pipelineResultsPageJSON(server.URL, runID, "res_1", "res_2"))
		case "/files/" + runID + "/res_1.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/files/" + runID + "/res_2.tar.gz":
			if archive2Requests.Add(1) == 1 {
				writePartialArchiveAndClose(t, w, archiveBytes[:10], len(archiveBytes))
				return
			}
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, _, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", runID,
		"--name", "pipeline-resume",
	)
	require.Error(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "pipeline-resume")
	metadata := mustLoadDownloadMetadata(t, runDir)
	require.NotNil(t, metadata.Pending)
	assert.Equal(t, []string{"res_2"}, metadata.Pending.ResultIDs)

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--name", "pipeline-resume",
	)
	require.NoError(t, err)

	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "files", "result.txt"))
	assert.FileExists(t, filepath.Join(runDir, "results", "res_2", "files", "result.txt"))
	metadata = mustLoadDownloadMetadata(t, runDir)
	assert.Nil(t, metadata.Pending)
}

func TestPipelineResultManifestAppenderSkipsPreviouslyIndexedResults(t *testing.T) {
	cwd := t.TempDir()
	runDir := filepath.Join(cwd, "runs", "append")
	resultDir := filepath.Join(runDir, "results", "res_1")
	require.NoError(t, os.MkdirAll(filepath.Join(resultDir, "files"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(resultDir, "archive.tar.gz"), []byte("archive"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(resultDir, "files", "metrics.json"), []byte("{}"), 0o644))
	require.NoError(t, savePipelineResultMetadata(resultDir, downloadPipelineResultInfo{
		ID:       "res_1",
		Metadata: map[string]any{"id": "res_1", "score": 0.9},
	}))

	manifest := newPipelineResultManifestAppender(runDir)
	require.NoError(t, manifest.appendResult(resultDir, "res_1"))
	require.NoError(t, manifest.appendResult(resultDir, "res_1"))

	entries := readDownloadResultsTestJSONL(t, filepath.Join(runDir, "results", "index.jsonl"))
	require.Len(t, entries, 1)
	assert.Equal(t, "res_1", entries[0]["id"])
	paths, ok := entries[0]["paths"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "results/res_1/archive.tar.gz", paths["archive"])
	assert.Equal(t, "results/res_1/files/metrics.json", paths["metrics"])
}

func TestAppendDownloadJSONLReturnsCloseErrors(t *testing.T) {
	closeErr := errors.New("close failed")
	writer := &closeErrorWriter{err: closeErr}

	err := appendDownloadJSONL(writer, map[string]any{"id": "res_1"})

	require.ErrorIs(t, err, closeErr)
	assert.JSONEq(t, `{"id":"res_1"}`, strings.TrimSpace(writer.String()))
}

func TestPipelineFailedStillDownloadsDiscoveredResults(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runID := "prot_des_123"
	archiveBytes := makeTarGzArchive(t, map[string]string{"result.txt": "failed-but-downloaded"})

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/protein/design/" + runID:
			writeJSON(t, w, pipelineGetResponseJSON(runID, "failed", "ws_123", "res_1", "bad_news"))
		case "/compute/v1/protein/design/" + runID + "/results":
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

	_, _, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", runID,
		"--name", "pipeline-failed",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad_news")

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "pipeline-failed")
	assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "files", "result.txt"))
}

func TestPipelineStoppedDownloadsAndSucceeds(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runID := "prot_des_123"
	archiveBytes := makeTarGzArchive(t, map[string]string{"result.txt": "stopped"})

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/protein/design/" + runID:
			writeJSON(t, w, pipelineGetResponseJSON(runID, "stopped", "ws_123", "res_1", ""))
		case "/compute/v1/protein/design/" + runID + "/results":
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

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", runID,
		"--name", "pipeline-stopped",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "pipeline-stopped")
	assert.Equal(t, runDir+"\n", stdout)
	assert.NotEmpty(t, stderr)
	assert.FileExists(t, filepath.Join(runDir, "results", "res_1", "files", "result.txt"))
}

func TestPipelineMissingCheckpointedResultFails(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	runDir := filepath.Join(cwd, "runs", "resume")
	require.NoError(t, os.MkdirAll(runDir, 0o755))

	runID := "prot_des_123"
	metadata := newDownloadRunMetadata("resume", downloadRunTypeProteinDesign, runID, nil)
	lastID := "res_1"
	metadata.Pending = &downloadPendingState{
		Kind:       downloadPendingKindResultPage,
		AfterID:    nil,
		PageLastID: &lastID,
		ResultIDs:  []string{"res_1"},
	}
	require.NoError(t, saveDownloadMetadata(runDir, metadata))

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/protein/design/" + runID:
			writeJSON(t, w, pipelineGetResponseJSON(runID, "succeeded", "ws_123", "res_1", ""))
		case "/compute/v1/protein/design/" + runID + "/results":
			writeJSON(t, w, pipelineResultsPageJSON(server.URL, runID, "res_2"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, _, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--root-dir", "runs",
		"--name", "resume",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "res_1 is no longer present")
}

func TestDownloadResultsVerboseWritesProgressToStderr(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var retrieveCalls atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			if retrieveCalls.Add(1) == 1 {
				writeJSON(t, w, predictionResponseJSON("pred_123", "running", "ws_123", "", ""))
				return
			}
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "verbose-run",
		"--progress-format", "text",
		"--verbose",
		"--poll-interval-seconds", "0",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "verbose-run")
	assert.Equal(t, runDir+"\n", stdout)
	assert.Contains(t, stderr, "Waiting for prediction")
	assert.Contains(t, stderr, "Downloading archive")
}

func TestDownloadResultsDefaultJSONLProgressWritesMachineReadableEvents(t *testing.T) {
	setDownloadResultsTestEnv(t)
	cwd := t.TempDir()
	t.Chdir(cwd)

	archiveBytes := makeTarGzArchive(t, map[string]string{"nested/output.txt": "done"})
	var retrieveCalls atomic.Int32

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compute/v1/predictions/structure-and-binding/pred_123":
			if retrieveCalls.Add(1) == 1 {
				writeJSON(t, w, predictionResponseJSON("pred_123", "running", "ws_123", "", ""))
				return
			}
			writeJSON(t, w, predictionResponseJSON("pred_123", "succeeded", "ws_123", server.URL+"/files/prediction.tar.gz", ""))
		case "/files/prediction.tar.gz":
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runDownloadResultsCLI(
		t,
		"--base-url", server.URL,
		"--api-key", "test-key",
		"download-results",
		"--id", "pred_123",
		"--name", "jsonl-run",
		"--poll-interval-seconds", "0",
	)
	require.NoError(t, err)

	runDir := filepath.Join(cwd, downloadResultsDefaultRootDir, "jsonl-run")
	assert.Equal(t, runDir+"\n", stdout)

	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	require.NotEmpty(t, lines)

	seenEvents := map[string]bool{}
	for _, line := range lines {
		var event map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &event))
		require.Equal(t, runDir, event["run_dir"])
		require.Equal(t, "jsonl-run", event["name"])
		_, hasTimestamp := event["ts"]
		require.True(t, hasTimestamp)

		eventName, ok := event["event"].(string)
		require.True(t, ok)
		seenEvents[eventName] = true
	}

	assert.True(t, seenEvents["waiting"])
	assert.True(t, seenEvents["status"])
	assert.True(t, seenEvents["archive_download"])
	assert.True(t, seenEvents["archive_extract"])
	assert.True(t, seenEvents["ready"])
}

func runDownloadResultsCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := newDownloadResultsTestRoot(&stdout, &stderr)
	err := root.Run(context.Background(), append([]string{"boltz-api"}, args...))
	return stdout.String(), stderr.String(), err
}

type closeErrorWriter struct {
	bytes.Buffer
	err error
}

func (w *closeErrorWriter) Close() error {
	return w.err
}

func newDownloadResultsTestRoot(stdout io.Writer, stderr io.Writer) *cli.Command {
	root := &cli.Command{
		Name:      "boltz-api",
		Writer:    stdout,
		ErrWriter: stderr,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url"},
			&cli.StringFlag{Name: "format", Value: "auto"},
			&cli.StringFlag{Name: "transform"},
			&cli.BoolFlag{Name: "raw-output"},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_API_KEY"),
			},
		},
	}
	ApplyCustomizations(root)
	return root
}

func setDownloadResultsTestEnv(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(payload))
}

func predictionResponseJSON(id string, status string, workspaceID string, archiveURL string, errorCode string) map[string]any {
	payload := map[string]any{
		"id":           id,
		"status":       status,
		"workspace_id": workspaceID,
		"started_at":   time.Now().UTC().Format(time.RFC3339),
		"completed_at": time.Now().UTC().Format(time.RFC3339),
		"input": map[string]any{
			"entities": []map[string]any{{
				"chain_ids": []string{"A"},
				"type":      "protein",
				"value":     "ACDE",
			}},
			"num_samples": 1,
		},
	}
	if errorCode != "" {
		payload["error"] = map[string]any{"code": errorCode}
	}
	if archiveURL != "" {
		payload["output"] = map[string]any{
			"archive": map[string]any{
				"url":            archiveURL,
				"url_expires_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			},
		}
	}
	return payload
}

func pipelineGetResponseJSON(id string, status string, workspaceID string, latestResultID string, errorCode string) map[string]any {
	payload := map[string]any{
		"id":           id,
		"status":       status,
		"workspace_id": workspaceID,
		"started_at":   time.Now().UTC().Format(time.RFC3339),
		"completed_at": time.Now().UTC().Format(time.RFC3339),
		"stopped_at":   time.Now().UTC().Format(time.RFC3339),
		"progress": map[string]any{
			"latest_result_id": latestResultID,
		},
	}
	if errorCode != "" {
		payload["error"] = map[string]any{"code": errorCode}
	}
	return payload
}

func pipelineResultsPageJSON(baseURL string, runID string, resultIDs ...string) map[string]any {
	data := make([]map[string]any, 0, len(resultIDs))
	for _, resultID := range resultIDs {
		result := pipelineResultJSON(runID, resultID)
		result["artifacts"] = map[string]any{
			"archive": map[string]any{
				"url":            fmt.Sprintf("%s/files/%s/%s.tar.gz", baseURL, runID, resultID),
				"url_expires_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			},
		}
		data = append(data, result)
	}
	var lastID any
	if len(resultIDs) > 0 {
		lastID = resultIDs[len(resultIDs)-1]
	}
	return map[string]any{
		"data":     data,
		"first_id": nil,
		"last_id":  lastID,
		"has_more": false,
	}
}

func pipelineResultJSON(runID string, resultID string) map[string]any {
	result := map[string]any{
		"id":         resultID,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"metrics": map[string]any{
			"binding_confidence":   0.91,
			"complex_iplddt":       0.82,
			"complex_plddt":        0.83,
			"helix_fraction":       0.10,
			"iptm":                 0.84,
			"loop_fraction":        0.20,
			"min_interaction_pae":  3.40,
			"optimization_score":   0.85,
			"ptm":                  0.86,
			"sheet_fraction":       0.30,
			"structure_confidence": 0.87,
		},
		"warnings": []map[string]any{{
			"code":    "low_confidence",
			"message": "check manually",
		}},
	}
	switch {
	case strings.HasPrefix(runID, "sm_scr_"):
		result["external_id"] = "mol-" + resultID
		result["smiles"] = "Cc1ccc(C(=O)NC)cc1"
	case strings.HasPrefix(runID, "sm_des_"):
		result["smiles"] = "N#Cc1ccccc1"
	case strings.HasPrefix(runID, "prot_scr_"):
		result["external_id"] = "protein-" + resultID
		result["entities"] = []map[string]any{{
			"chain_ids": []string{"A"},
			"type":      "protein",
			"value":     "ACDE",
		}}
	case strings.HasPrefix(runID, "prot_des_"):
		result["entities"] = []map[string]any{{
			"chain_ids": []string{"A"},
			"type":      "protein",
			"value":     "WXYZ",
		}}
	}
	return result
}

func emptyPipelinePageJSON() map[string]any {
	return map[string]any{
		"data":     []any{},
		"first_id": nil,
		"last_id":  nil,
		"has_more": false,
	}
}

func assertPipelineResultMetadata(t *testing.T, runID string, metadata map[string]any) {
	t.Helper()

	_, hasMetrics := metadata["metrics"].(map[string]any)
	assert.True(t, hasMetrics)
	_, hasWarnings := metadata["warnings"].([]any)
	assert.True(t, hasWarnings)

	switch {
	case strings.HasPrefix(runID, "sm_scr_"):
		assert.Equal(t, "mol-res_1", metadata["external_id"])
		assert.Equal(t, "Cc1ccc(C(=O)NC)cc1", metadata["smiles"])
	case strings.HasPrefix(runID, "sm_des_"):
		assert.Equal(t, "N#Cc1ccccc1", metadata["smiles"])
	case strings.HasPrefix(runID, "prot_scr_"):
		assert.Equal(t, "protein-res_1", metadata["external_id"])
		entities, ok := metadata["entities"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, entities)
	case strings.HasPrefix(runID, "prot_des_"):
		entities, ok := metadata["entities"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, entities)
	}
}

func readDownloadResultsTestJSON(t *testing.T, path string) map[string]any {
	t.Helper()

	body, err := os.ReadFile(path)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload
}

func readDownloadResultsTestJSONL(t *testing.T, path string) []map[string]any {
	t.Helper()

	body, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	results := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var payload map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &payload))
		results = append(results, payload)
	}
	return results
}

func makeTarGzArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	for name, contents := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(contents)),
		}
		require.NoError(t, tarWriter.WriteHeader(header))
		_, err := tarWriter.Write([]byte(contents))
		require.NoError(t, err)
	}

	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())
	return buffer.Bytes()
}

func writePartialArchiveAndClose(t *testing.T, w http.ResponseWriter, body []byte, totalLength int) {
	t.Helper()

	hijacker, ok := w.(http.Hijacker)
	require.True(t, ok)

	conn, buffer, err := hijacker.Hijack()
	require.NoError(t, err)
	defer conn.Close()

	_, err = fmt.Fprintf(buffer, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: application/octet-stream\r\n\r\n", totalLength)
	require.NoError(t, err)
	_, err = buffer.Write(body)
	require.NoError(t, err)
	require.NoError(t, buffer.Flush())
}

func mustLoadDownloadMetadata(t *testing.T, runDir string) downloadRunMetadata {
	t.Helper()
	metadata, err := loadDownloadMetadata(runDir)
	require.NoError(t, err)
	return metadata
}
