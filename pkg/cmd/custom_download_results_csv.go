// Custom CLI extension code. Not generated.
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const downloadResultsSummaryCSVName = "summary.csv"

// Dropped from the summary because they add noise (paths are local-machine,
// created_at duplicates row order). Both stay in results/index.jsonl.
var smallMoleculeSummaryDroppedKeys = []string{"paths", "created_at"}

// Only small-molecule pipelines emit rows that flatten cleanly into a table
// (SMILES + ADME + metrics). Protein pipelines store sequences inside an
// `entities` array that doesn't fit a single tabular row.
func isSmallMoleculePipelineRunType(runType downloadRunType) bool {
	switch runType {
	case downloadRunTypeSmallMoleculeDesign, downloadRunTypeSmallMoleculeScreen:
		return true
	default:
		return false
	}
}

// generateSmallMoleculeSummary builds results/summary.csv as a fresh
// projection of results/index.jsonl: drops smallMoleculeSummaryDroppedKeys,
// flattens nested maps with dotted keys, encodes slices as compact JSON.
// Columns: smiles, id, then sorted remainder. Removes a stale summary if the
// JSONL is gone so the sidecar never outlives its source.
func generateSmallMoleculeSummary(runDir string) error {
	resultsDir := filepath.Join(runDir, "results")
	jsonlPath := filepath.Join(resultsDir, downloadResultsResultIndexName)
	csvPath := filepath.Join(resultsDir, downloadResultsSummaryCSVName)

	rows, err := readDownloadJSONLRows(jsonlPath)
	if errors.Is(err, os.ErrNotExist) {
		return removeDownloadFileIfExists(csvPath)
	}
	if err != nil {
		return err
	}

	flattened := make([]map[string]string, 0, len(rows))
	keySet := map[string]struct{}{}
	for _, row := range rows {
		flat := flattenSmallMoleculeSummaryRow(row)
		flattened = append(flattened, flat)
		for key := range flat {
			keySet[key] = struct{}{}
		}
	}

	headers := orderedSmallMoleculeSummaryHeaders(keySet)
	return writeDownloadCSVFile(csvPath, headers, flattened)
}

// appendSmallMoleculeSummaryRow appends one record to summary.csv using the
// caller-supplied header order. Missing keys become empty cells. This is the
// fast path used when the row's columns are already a subset of the file's
// header line — no need to rewrite the whole file.
func appendSmallMoleculeSummaryRow(csvPath string, headerOrder []string, flat map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(csvPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	record := make([]string, len(headerOrder))
	for i, header := range headerOrder {
		record[i] = flat[header]
	}
	writer := csv.NewWriter(file)
	if err := writer.Write(record); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

// readSmallMoleculeSummaryHeaders returns the column names from the CSV's
// first line, or nil if the file is missing. Used to seed the appender's
// cache so subsequent rows can take the fast-path.
func readSmallMoleculeSummaryHeaders(csvPath string) ([]string, error) {
	file, err := os.Open(csvPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	record, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func flattenSmallMoleculeSummaryRow(row map[string]any) map[string]string {
	cloned := cloneDownloadJSONMap(row)
	for _, key := range smallMoleculeSummaryDroppedKeys {
		delete(cloned, key)
	}
	flat := map[string]string{}
	flattenManifestValue("", cloned, flat)
	return flat
}

func readDownloadJSONLRows(path string) ([]map[string]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows := []map[string]any{}
	decoder := json.NewDecoder(file)
	for {
		var row map[string]any
		if err := decoder.Decode(&row); errors.Is(err, io.EOF) {
			return rows, nil
		} else if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
}

// Recursively flattens nested maps with dotted keys; slices and other
// structured values are JSON-encoded so every CSV cell stays a single scalar.
// Empty maps become "{}" so consumers can distinguish them from missing keys.
func flattenManifestValue(prefix string, value any, out map[string]string) {
	switch v := value.(type) {
	case nil:
		out[prefix] = ""
	case string:
		out[prefix] = v
	case bool:
		out[prefix] = strconv.FormatBool(v)
	case float64:
		out[prefix] = strconv.FormatFloat(v, 'g', -1, 64)
	case json.Number:
		out[prefix] = v.String()
	case map[string]any:
		if len(v) == 0 {
			out[prefix] = "{}"
			return
		}
		for k, sub := range v {
			childKey := k
			if prefix != "" {
				childKey = prefix + "." + k
			}
			flattenManifestValue(childKey, sub, out)
		}
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			out[prefix] = fmt.Sprint(v)
			return
		}
		out[prefix] = string(encoded)
	}
}

// "smiles" leads (the scientific identifier readers scan first), then "id",
// then sorted remainder for diff-friendliness.
func orderedSmallMoleculeSummaryHeaders(keySet map[string]struct{}) []string {
	rest := make([]string, 0, len(keySet))
	hasID := false
	hasSMILES := false
	for key := range keySet {
		switch key {
		case "id":
			hasID = true
		case "smiles":
			hasSMILES = true
		default:
			rest = append(rest, key)
		}
	}
	sort.Strings(rest)

	headers := make([]string, 0, len(keySet))
	if hasSMILES {
		headers = append(headers, "smiles")
	}
	if hasID {
		headers = append(headers, "id")
	}
	return append(headers, rest...)
}

func writeDownloadCSVFile(path string, headers []string, rows []map[string]string) error {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return err
		}
	}
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, header := range headers {
			record[i] = row[header]
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	return writeDownloadFileAtomically(path, []byte(builder.String()))
}

func removeDownloadFileIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
