package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpOutputIncludesRepeatableArrayGuidance(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	smallMolecule := runCLI(t, binary, env, "small-molecule:library-screen", "start", "--help")
	require.Equal(t, 0, smallMolecule.ExitCode, smallMolecule.Stderr)
	require.Contains(t, smallMolecule.Stdout, "Repeat --molecule to add entries.")
	require.Contains(t, smallMolecule.Stdout, "use the body field molecules")

	protein := runCLI(t, binary, env, "protein:library-screen", "start", "--help")
	require.Equal(t, 0, protein.ExitCode, protein.Stderr)
	require.Contains(t, protein.Stdout, "Repeat --protein to add entries.")
	require.Contains(t, protein.Stdout, "use the body field proteins")
}

func TestStructuredIncludeObjectFlagSubprocess(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.json")
	require.NoError(t, os.WriteFile(inputPath, []byte(`{"entities":[{"type":"protein","value":"SEQ","modifications":[]}]}`), 0644))

	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	result := runCLI(
		t,
		binary,
		env,
		"--api-key", "test-key",
		"--base-url", server.URL,
		"--format", "raw",
		"predictions:structure-and-binding", "estimate-cost",
		"--input", "@json://"+inputPath,
		"--model", "boltz-2.1",
	)
	require.Equal(t, 0, result.ExitCode, result.Stderr)

	var body map[string]any
	require.NoError(t, json.Unmarshal(capturedBody, &body))

	input, ok := body["input"].(map[string]any)
	require.True(t, ok)
	entities, ok := input["entities"].([]any)
	require.True(t, ok)
	require.Len(t, entities, 1)
	entity, ok := entities[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "protein", entity["type"])
	require.Equal(t, "SEQ", entity["value"])
}

func TestAdmeEstimateCostSubprocess(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	var capturedMethod string
	var capturedPath string
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		capturedBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"cost":1,"billable_cost":1,"breakdown":{"application":"adme","cost":1,"billable_cost":1}}`))
	}))
	defer server.Close()

	result := runCLI(
		t,
		binary,
		env,
		"--api-key", "test-key",
		"--base-url", server.URL,
		"--format", "raw",
		"predictions:adme", "estimate-cost",
		"--input.molecules", `[{smiles: CCO, id: ethanol}]`,
		"--model", "adme-v1",
		"--workspace-id", "workspace-from-flag",
		"--idempotency-key", "idempotency-from-flag",
	)
	require.Equal(t, 0, result.ExitCode, result.Stderr)
	require.Equal(t, http.MethodPost, capturedMethod)
	require.Equal(t, "/compute/v1/predictions/adme/estimate-cost", capturedPath)

	var body map[string]any
	require.NoError(t, json.Unmarshal(capturedBody, &body))
	require.Equal(t, "adme-v1", body["model"])
	require.Equal(t, "workspace-from-flag", body["workspace_id"])
	require.Equal(t, "idempotency-from-flag", body["idempotency_key"])

	input, ok := body["input"].(map[string]any)
	require.True(t, ok)
	molecules, ok := input["molecules"].([]any)
	require.True(t, ok)
	require.Len(t, molecules, 1)
	molecule, ok := molecules[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "CCO", molecule["smiles"])
	require.Equal(t, "ethanol", molecule["id"])
}

func TestMergedInputFlagSubprocess(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	testCases := []struct {
		name       string
		command    []string
		inputBody  map[string]any
		assertBody func(*testing.T, map[string]any)
	}{
		{
			name:    "small-molecule-design-estimate-cost",
			command: []string{"small-molecule:design", "estimate-cost"},
			inputBody: map[string]any{
				"num_molecules":    10,
				"target":           map[string]any{"marker": "from-input"},
				"chemical_space":   "enamine_real",
				"molecule_filters": map[string]any{"boltz_smarts_catalog_filter_level": "recommended"},
				"workspace_id":     "workspace-from-input",
				"idempotency_key":  "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				require.Equal(t, float64(10), body["num_molecules"])
				require.Equal(t, "enamine_real", body["chemical_space"])
				filters, ok := body["molecule_filters"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "recommended", filters["boltz_smarts_catalog_filter_level"])
			},
		},
		{
			name:    "small-molecule-design-start",
			command: []string{"small-molecule:design", "start"},
			inputBody: map[string]any{
				"num_molecules":    10,
				"target":           map[string]any{"marker": "from-input"},
				"chemical_space":   "enamine_real",
				"molecule_filters": map[string]any{"boltz_smarts_catalog_filter_level": "recommended"},
				"workspace_id":     "workspace-from-input",
				"idempotency_key":  "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				require.Equal(t, float64(10), body["num_molecules"])
				require.Equal(t, "enamine_real", body["chemical_space"])
				filters, ok := body["molecule_filters"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "recommended", filters["boltz_smarts_catalog_filter_level"])
			},
		},
		{
			name:    "small-molecule-library-screen-estimate-cost",
			command: []string{"small-molecule:library-screen", "estimate-cost"},
			inputBody: map[string]any{
				"molecules":        []map[string]any{{"smiles": "CCO", "id": "mol-1"}},
				"target":           map[string]any{"marker": "from-input"},
				"molecule_filters": map[string]any{"boltz_smarts_catalog_filter_level": "recommended"},
				"workspace_id":     "workspace-from-input",
				"idempotency_key":  "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				molecules, ok := body["molecules"].([]any)
				require.True(t, ok)
				require.Len(t, molecules, 1)
				molecule, ok := molecules[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "mol-1", molecule["id"])
			},
		},
		{
			name:    "small-molecule-library-screen-start",
			command: []string{"small-molecule:library-screen", "start"},
			inputBody: map[string]any{
				"molecules":        []map[string]any{{"smiles": "CCO", "id": "mol-1"}},
				"target":           map[string]any{"marker": "from-input"},
				"molecule_filters": map[string]any{"boltz_smarts_catalog_filter_level": "recommended"},
				"workspace_id":     "workspace-from-input",
				"idempotency_key":  "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				molecules, ok := body["molecules"].([]any)
				require.True(t, ok)
				require.Len(t, molecules, 1)
				molecule, ok := molecules[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "mol-1", molecule["id"])
			},
		},
		{
			name:    "protein-design-estimate-cost",
			command: []string{"protein:design", "estimate-cost"},
			inputBody: map[string]any{
				"binder_specification": map[string]any{"modality": "peptide"},
				"num_proteins":         10,
				"target":               map[string]any{"marker": "from-input"},
				"workspace_id":         "workspace-from-input",
				"idempotency_key":      "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				require.Equal(t, float64(10), body["num_proteins"])
				binder, ok := body["binder_specification"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "peptide", binder["modality"])
			},
		},
		{
			name:    "protein-design-start",
			command: []string{"protein:design", "start"},
			inputBody: map[string]any{
				"binder_specification": map[string]any{"modality": "peptide"},
				"num_proteins":         10,
				"target":               map[string]any{"marker": "from-input"},
				"workspace_id":         "workspace-from-input",
				"idempotency_key":      "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				require.Equal(t, float64(10), body["num_proteins"])
				binder, ok := body["binder_specification"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "peptide", binder["modality"])
			},
		},
		{
			name:    "protein-library-screen-estimate-cost",
			command: []string{"protein:library-screen", "estimate-cost"},
			inputBody: map[string]any{
				"proteins":        []map[string]any{{"id": "prot-1", "entities": []map[string]any{{"type": "protein", "value": "SEQ"}}}},
				"target":          map[string]any{"marker": "from-input"},
				"workspace_id":    "workspace-from-input",
				"idempotency_key": "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				proteins, ok := body["proteins"].([]any)
				require.True(t, ok)
				require.Len(t, proteins, 1)
				protein, ok := proteins[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "prot-1", protein["id"])
			},
		},
		{
			name:    "protein-library-screen-start",
			command: []string{"protein:library-screen", "start"},
			inputBody: map[string]any{
				"proteins":        []map[string]any{{"id": "prot-1", "entities": []map[string]any{{"type": "protein", "value": "SEQ"}}}},
				"target":          map[string]any{"marker": "from-input"},
				"workspace_id":    "workspace-from-input",
				"idempotency_key": "idempotency-from-input",
			},
			assertBody: func(t *testing.T, body map[string]any) {
				proteins, ok := body["proteins"].([]any)
				require.True(t, ok)
				require.Len(t, proteins, 1)
				protein, ok := proteins[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "prot-1", protein["id"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputPath := filepath.Join(tmpDir, "input.json")
			targetPath := filepath.Join(tmpDir, "target.json")

			inputBytes, err := json.Marshal(tc.inputBody)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(inputPath, inputBytes, 0644))
			require.NoError(t, os.WriteFile(targetPath, []byte(`{"marker":"from-flag"}`), 0644))

			var capturedBody []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var readErr error
				capturedBody, readErr = io.ReadAll(r.Body)
				require.NoError(t, readErr)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()

			args := []string{
				"--api-key", "test-key",
				"--base-url", server.URL,
				"--format", "raw",
			}
			args = append(args, tc.command...)
			args = append(args,
				"--input", "@json://"+inputPath,
				"--target", "@json://"+targetPath,
				"--workspace-id", "workspace-from-flag",
				"--idempotency-key", "idempotency-from-flag",
			)

			result := runCLI(t, binary, env, args...)
			require.Equal(t, 0, result.ExitCode, result.Stderr)

			var body map[string]any
			require.NoError(t, json.Unmarshal(capturedBody, &body))

			require.Equal(t, "workspace-from-flag", body["workspace_id"])
			require.Equal(t, "idempotency-from-flag", body["idempotency_key"])

			target, ok := body["target"].(map[string]any)
			require.True(t, ok)
			require.Equal(t, "from-flag", target["marker"])

			tc.assertBody(t, body)
		})
	}
}

func TestNativeIDFlagParsesOnRetrieveCommands(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	result := runCLI(
		t,
		binary,
		env,
		"--api-key", "test-key",
		"--base-url", server.URL,
		"--format", "raw",
		"small-molecule:design", "retrieve",
		"--id", "run_123",
	)
	require.Equal(t, 0, result.ExitCode, result.Stderr)
	require.NotEmpty(t, requestPath)
	require.Contains(t, requestPath, "run_123")
	require.NotContains(t, result.Stderr, "flag provided but not defined")
}

func TestLegacyRunAndScreenIDFlagsAreRejected(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	runIDResult := runCLI(
		t,
		binary,
		env,
		"small-molecule:design", "retrieve",
		"--run-id", "run_123",
	)
	require.NotEqual(t, 0, runIDResult.ExitCode)
	require.Contains(t, runIDResult.Stderr, "flag provided but not defined")

	screenIDResult := runCLI(
		t,
		binary,
		env,
		"small-molecule:library-screen", "list-results",
		"--screen-id", "sm_scr_123",
	)
	require.NotEqual(t, 0, screenIDResult.ExitCode)
	require.Contains(t, screenIDResult.Stderr, "flag provided but not defined")
}
