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

const downloadResultsResultCSVName = "index.csv"

// writeDownloadCSVFromJSONL derives results/index.csv from results/index.jsonl.
// The CSV is a strict, deterministic projection of the JSONL: same rows in the
// same order, with nested maps flattened using dotted keys and slices encoded
// as compact JSON strings. The column set is the union of keys across rows;
// missing cells are empty. The "id" column is placed first and the remaining
// columns are sorted lexicographically so the file is diff-friendly.
//
// If the JSONL is missing, any existing CSV is removed so a stale sidecar
// never outlives the manifest it derives from.
func writeDownloadCSVFromJSONL(runDir string) error {
	resultsDir := filepath.Join(runDir, "results")
	jsonlPath := filepath.Join(resultsDir, downloadResultsResultIndexName)
	csvPath := filepath.Join(resultsDir, downloadResultsResultCSVName)

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
		flat := map[string]string{}
		flattenManifestValue("", row, flat)
		flattened = append(flattened, flat)
		for key := range flat {
			keySet[key] = struct{}{}
		}
	}

	headers := orderedManifestCSVHeaders(keySet)
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

func orderedManifestCSVHeaders(keySet map[string]struct{}) []string {
	headers := make([]string, 0, len(keySet))
	for key := range keySet {
		headers = append(headers, key)
	}
	sort.Strings(headers)
	// Pin "id" first when present; it is the natural row identifier and
	// almost always the first column a reader wants to see.
	for i, header := range headers {
		if header == "id" {
			headers = append([]string{"id"}, append(headers[:i], headers[i+1:]...)...)
			break
		}
	}
	return headers
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
