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

func TestWriteSmallMoleculeSummaryCSVFlattenAndDropPaths(t *testing.T) {
	runDir := t.TempDir()
	jsonlPath := filepath.Join(runDir, "results", downloadResultsResultIndexName)
	require.NoError(t, os.MkdirAll(filepath.Dir(jsonlPath), 0o755))

	writeJSONL(t, jsonlPath, []string{
		`{"id":"res_1","smiles":"CCO","created_at":"2026-06-07T19:48:31.116Z","adme":{"solubility":"high"},"metrics":{"iptm":0.91},"paths":{"archive":"results/res_1/archive.tar.gz","files":"results/res_1/files"}}`,
		`{"id":"res_2","smiles":"CCN","created_at":"2026-06-07T19:48:55.852Z","adme":{"solubility":"low"},"metrics":{"iptm":0.78,"binding_confidence":0.42},"paths":{"archive":"results/res_2/archive.tar.gz"}}`,
	})

	require.NoError(t, writeSmallMoleculeSummaryCSV(runDir))

	headers, rows := readDownloadResultsTestCSV(t, filepath.Join(runDir, "results", downloadResultsSummaryCSVName))

	// smiles first, id second, rest alphabetical; no paths.* or created_at columns.
	assert.Equal(t, []string{
		"smiles",
		"id",
		"adme.solubility",
		"metrics.binding_confidence",
		"metrics.iptm",
	}, headers)
	assert.NotContains(t, headers, "created_at")

	require.Len(t, rows, 2)
	assert.Equal(t, "res_1", rows[0]["id"])
	assert.Equal(t, "CCO", rows[0]["smiles"])
	assert.Equal(t, "high", rows[0]["adme.solubility"])
	assert.Equal(t, "0.91", rows[0]["metrics.iptm"])
	// row 1 lacks binding_confidence; cell must be empty rather than missing.
	assert.Equal(t, "", rows[0]["metrics.binding_confidence"])

	assert.Equal(t, "res_2", rows[1]["id"])
	assert.Equal(t, "CCN", rows[1]["smiles"])
	assert.Equal(t, "0.42", rows[1]["metrics.binding_confidence"])

	// Sanity: no paths.* leaked into the CSV.
	for _, header := range headers {
		assert.False(t, strings.HasPrefix(header, "paths."), "paths.* column should not appear: %s", header)
	}
}

func TestWriteSmallMoleculeSummaryCSVRemovesStaleCSVWhenJSONLMissing(t *testing.T) {
	runDir := t.TempDir()
	csvPath := filepath.Join(runDir, "results", downloadResultsSummaryCSVName)
	require.NoError(t, os.MkdirAll(filepath.Dir(csvPath), 0o755))
	require.NoError(t, os.WriteFile(csvPath, []byte("id\nold\n"), 0o644))

	require.NoError(t, writeSmallMoleculeSummaryCSV(runDir))

	_, err := os.Stat(csvPath)
	assert.True(t, os.IsNotExist(err), "expected stale summary.csv to be removed, got err=%v", err)
}

func TestWriteSmallMoleculeSummaryCSVNoopWhenNoResultsDir(t *testing.T) {
	runDir := t.TempDir()
	// No results/ directory yet. The writer must not fail when there is
	// nothing to derive from: appendEntry calls this before the directory has
	// been populated on the very first row.
	require.NoError(t, writeSmallMoleculeSummaryCSV(runDir))
}

func TestPipelineResultManifestAppenderWritesSummaryForSmallMolecule(t *testing.T) {
	runDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(runDir, "results"), 0o755))

	manifest := newPipelineResultManifestAppender(runDir, downloadRunTypeSmallMoleculeDesign)
	require.NoError(t, manifest.appendMetadata(map[string]any{
		"id":     "res_1",
		"smiles": "CCO",
		"adme":   map[string]any{"solubility": "high"},
		"paths": map[string]any{
			"archive": "results/res_1/archive.tar.gz",
		},
	}, "res_1"))

	headers, rows := readDownloadResultsTestCSV(t, filepath.Join(runDir, "results", downloadResultsSummaryCSVName))
	assert.Equal(t, []string{"smiles", "id", "adme.solubility"}, headers)
	require.Len(t, rows, 1)
	assert.Equal(t, "res_1", rows[0]["id"])
	assert.Equal(t, "CCO", rows[0]["smiles"])
	assert.Equal(t, "high", rows[0]["adme.solubility"])
}

func TestPipelineResultManifestAppenderSkipsSummaryForProteinDesign(t *testing.T) {
	runDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(runDir, "results"), 0o755))

	manifest := newPipelineResultManifestAppender(runDir, downloadRunTypeProteinDesign)
	require.NoError(t, manifest.appendMetadata(map[string]any{
		"id": "res_1",
		"entities": []any{
			map[string]any{"type": "protein", "value": "MKTII"},
		},
	}, "res_1"))

	// JSONL is written for every pipeline...
	assert.FileExists(t, filepath.Join(runDir, "results", downloadResultsResultIndexName))
	// ...but the summary.csv is small-molecule only.
	_, err := os.Stat(filepath.Join(runDir, "results", downloadResultsSummaryCSVName))
	assert.True(t, os.IsNotExist(err), "summary.csv should not be written for non-small-molecule runs")
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
