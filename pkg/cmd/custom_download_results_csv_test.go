// Custom CLI extension code. Not generated.
package cmd

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteDownloadCSVFromJSONLBuildsFlatRows(t *testing.T) {
	runDir := t.TempDir()
	jsonlPath := filepath.Join(runDir, "results", downloadResultsResultIndexName)
	require.NoError(t, os.MkdirAll(filepath.Dir(jsonlPath), 0o755))

	writeJSONL(t, jsonlPath, []string{
		`{"id":"res_1","name":"binder-a","paths":{"archive":"results/res_1/archive.tar.gz","files":"results/res_1/files"},"metrics":{"plddt":0.91,"iptm":0.78}}`,
		`{"id":"res_2","name":"binder-b","paths":{"archive":"results/res_2/archive.tar.gz"},"metrics":{"plddt":0.87}}`,
	})

	require.NoError(t, writeDownloadCSVFromJSONL(runDir))

	headers, rows := readDownloadResultsTestCSV(t, filepath.Join(runDir, "results", downloadResultsResultCSVName))

	// id must be the first column; remaining columns are sorted lexicographically.
	require.Equal(t, "id", headers[0])
	assert.Equal(t, []string{
		"id",
		"metrics.iptm",
		"metrics.plddt",
		"name",
		"paths.archive",
		"paths.files",
	}, headers)

	require.Len(t, rows, 2)
	assert.Equal(t, "res_1", rows[0]["id"])
	assert.Equal(t, "binder-a", rows[0]["name"])
	assert.Equal(t, "results/res_1/archive.tar.gz", rows[0]["paths.archive"])
	assert.Equal(t, "results/res_1/files", rows[0]["paths.files"])
	assert.Equal(t, "0.91", rows[0]["metrics.plddt"])
	assert.Equal(t, "0.78", rows[0]["metrics.iptm"])

	// Row 2 lacks paths.files and metrics.iptm; cells must be empty rather than missing.
	assert.Equal(t, "res_2", rows[1]["id"])
	assert.Equal(t, "", rows[1]["paths.files"])
	assert.Equal(t, "", rows[1]["metrics.iptm"])
	assert.Equal(t, "0.87", rows[1]["metrics.plddt"])
}

func TestWriteDownloadCSVFromJSONLEncodesSlicesAsJSON(t *testing.T) {
	runDir := t.TempDir()
	jsonlPath := filepath.Join(runDir, "results", downloadResultsResultIndexName)
	require.NoError(t, os.MkdirAll(filepath.Dir(jsonlPath), 0o755))

	writeJSONL(t, jsonlPath, []string{
		`{"id":"res_1","chains":["A","B"],"flags":{"ok":true,"label":""}}`,
	})

	require.NoError(t, writeDownloadCSVFromJSONL(runDir))

	headers, rows := readDownloadResultsTestCSV(t, filepath.Join(runDir, "results", downloadResultsResultCSVName))
	require.Equal(t, []string{"id", "chains", "flags.label", "flags.ok"}, headers)
	require.Len(t, rows, 1)

	assert.Equal(t, `["A","B"]`, rows[0]["chains"])
	assert.Equal(t, "true", rows[0]["flags.ok"])
	assert.Equal(t, "", rows[0]["flags.label"])
}

func TestWriteDownloadCSVFromJSONLRemovesStaleCSVWhenJSONLMissing(t *testing.T) {
	runDir := t.TempDir()
	csvPath := filepath.Join(runDir, "results", downloadResultsResultCSVName)
	require.NoError(t, os.MkdirAll(filepath.Dir(csvPath), 0o755))
	require.NoError(t, os.WriteFile(csvPath, []byte("id\nold\n"), 0o644))

	require.NoError(t, writeDownloadCSVFromJSONL(runDir))

	_, err := os.Stat(csvPath)
	assert.True(t, os.IsNotExist(err), "expected stale CSV to be removed, got err=%v", err)
}

func TestWriteDownloadCSVFromJSONLNoopWhenNoResultsDir(t *testing.T) {
	runDir := t.TempDir()
	// No results/ directory yet. The CSV writer must not fail when there is
	// nothing to derive from: appendEntry calls this before the directory has
	// been populated on the very first row.
	require.NoError(t, writeDownloadCSVFromJSONL(runDir))
}

func TestPipelineResultManifestAppenderWritesCSVAlongsideJSONL(t *testing.T) {
	runDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(runDir, "results"), 0o755))

	manifest := newPipelineResultManifestAppender(runDir)
	require.NoError(t, manifest.appendMetadata(map[string]any{
		"id":   "res_1",
		"name": "binder-a",
		"paths": map[string]any{
			"archive": "results/res_1/archive.tar.gz",
		},
	}, "res_1"))
	require.NoError(t, manifest.appendMetadata(map[string]any{
		"id":   "res_2",
		"name": "binder-b",
	}, "res_2"))

	headers, rows := readDownloadResultsTestCSV(t, filepath.Join(runDir, "results", downloadResultsResultCSVName))
	require.Equal(t, []string{"id", "name", "paths.archive"}, headers)
	require.Len(t, rows, 2)
	assert.Equal(t, "res_1", rows[0]["id"])
	assert.Equal(t, "results/res_1/archive.tar.gz", rows[0]["paths.archive"])
	assert.Equal(t, "res_2", rows[1]["id"])
	// res_2 has no paths.archive; we still want a header column for it because
	// res_1 had one, and the cell must be empty for res_2.
	assert.Equal(t, "", rows[1]["paths.archive"])
}

func writeJSONL(t *testing.T, path string, lines []string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644))
}

func readDownloadResultsTestCSV(t *testing.T, path string) ([]string, []map[string]string) {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.NotEmpty(t, records, "expected at least a header row in %s", path)

	headers := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := map[string]string{}
		for i, value := range record {
			row[headers[i]] = value
		}
		rows = append(rows, row)
	}
	return headers, rows
}
