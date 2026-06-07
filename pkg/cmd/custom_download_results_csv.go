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

// isSmallMoleculePipelineRunType reports whether the run produces tabular
// per-molecule rows (SMILES + scores) that make sense as a CSV summary.
// Other pipelines (protein design / screen) emit per-row data that does not
// flatten cleanly to a single table, so we skip the summary for them.
func isSmallMoleculePipelineRunType(runType downloadRunType) bool {
	switch runType {
	case downloadRunTypeSmallMoleculeDesign, downloadRunTypeSmallMoleculeScreen:
		return true
	default:
		return false
	}
}

// writeSmallMoleculeSummaryCSV derives results/summary.csv from
// results/index.jsonl for small-molecule pipelines. The CSV is a strict,
// deterministic projection of the JSONL with two transformations:
//
//   - the per-row "paths" object is dropped (local file pointers are noise in
//     a scientist-facing summary)
//   - remaining nested maps are flattened with dotted keys; slices are encoded
//     as compact JSON strings
//
// The column set is the union of keys across rows; missing cells are empty.
// The "smiles" column is placed first (the scientific identifier the
// reader is most likely to scan), "id" second, and the remaining columns
// are sorted lexicographically.
//
// If the JSONL is missing, any existing summary is removed so a stale sidecar
// never outlives the manifest it derives from.
func writeSmallMoleculeSummaryCSV(runDir string) error {
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
		cloned := cloneDownloadJSONMap(row)
		delete(cloned, "paths")

		flat := map[string]string{}
		flattenManifestValue("", cloned, flat)
		flattened = append(flattened, flat)
		for key := range flat {
			keySet[key] = struct{}{}
		}
	}

	headers := orderedSummaryCSVHeaders(keySet)
	return writeDownloadCSVFile(csvPath, headers, flattened)
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

// flattenManifestValue recursively flattens nested maps with dotted keys.
// Slices and unknown structured values are written as compact JSON strings so
// every CSV cell is a single scalar. Empty maps become "{}" so a downstream
// reader can tell them apart from a missing key.
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

// orderedSummaryCSVHeaders pins the natural identifier columns first
// ("smiles" then "id") and sorts the rest lexicographically so the file is
// diff-friendly and easy to scan.
func orderedSummaryCSVHeaders(keySet map[string]struct{}) []string {
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
