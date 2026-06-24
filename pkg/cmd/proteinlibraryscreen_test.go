// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestProteinLibraryScreenRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "retrieve",
			"--id", "id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinLibraryScreenList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinLibraryScreenDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "delete-data",
			"--id", "id",
		)
	})
}

func TestProteinLibraryScreenEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "estimate-cost",
			"--protein", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], id: id}",
			"--target", "{chain_selection: {A: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], epitope_residues: [10, 11, 12], flexible_residues: [5, 6, 7], non_binding_residues: [0, 1, 2]}}, structure: {type: url, url: https://example.com}, type: structure_template}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(proteinLibraryScreenEstimateCost)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "estimate-cost",
			"--protein.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--protein.id", "id",
			"--target", "{chain_selection: {A: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], epitope_residues: [10, 11, 12], flexible_residues: [5, 6, 7], non_binding_residues: [0, 1, 2]}}, structure: {type: url, url: https://example.com}, type: structure_template}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"proteins:\n" +
			"  - entities:\n" +
			"      - chain_ids:\n" +
			"          - string\n" +
			"        type: protein\n" +
			"        value: value\n" +
			"        cyclic: true\n" +
			"        modifications:\n" +
			"          - residue_index: 0\n" +
			"            type: ccd\n" +
			"            value: value\n" +
			"    id: id\n" +
			"target:\n" +
			"  chain_selection:\n" +
			"    A:\n" +
			"      chain_type: polymer\n" +
			"      crop_residues:\n" +
			"        - 0\n" +
			"        - 1\n" +
			"        - 2\n" +
			"        - 3\n" +
			"        - 4\n" +
			"        - 5\n" +
			"        - 6\n" +
			"        - 7\n" +
			"        - 8\n" +
			"        - 9\n" +
			"        - 10\n" +
			"        - 11\n" +
			"        - 12\n" +
			"      epitope_residues:\n" +
			"        - 10\n" +
			"        - 11\n" +
			"        - 12\n" +
			"      flexible_residues:\n" +
			"        - 5\n" +
			"        - 6\n" +
			"        - 7\n" +
			"      non_binding_residues:\n" +
			"        - 0\n" +
			"        - 1\n" +
			"        - 2\n" +
			"  structure:\n" +
			"    type: url\n" +
			"    url: https://example.com\n" +
			"  type: structure_template\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"protein:library-screen", "estimate-cost",
		)
	})
}

func TestProteinLibraryScreenListResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "list-results",
			"--max-items", "10",
			"--id", "id",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--ids", "ids",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinLibraryScreenResume(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "resume",
			"--id", "id",
		)
	})
}

func TestProteinLibraryScreenStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "start",
			"--protein", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], id: id}",
			"--target", "{chain_selection: {A: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], epitope_residues: [10, 11, 12], flexible_residues: [5, 6, 7], non_binding_residues: [0, 1, 2]}}, structure: {type: url, url: https://example.com}, type: structure_template}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(proteinLibraryScreenStart)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "start",
			"--protein.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--protein.id", "id",
			"--target", "{chain_selection: {A: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], epitope_residues: [10, 11, 12], flexible_residues: [5, 6, 7], non_binding_residues: [0, 1, 2]}}, structure: {type: url, url: https://example.com}, type: structure_template}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"proteins:\n" +
			"  - entities:\n" +
			"      - chain_ids:\n" +
			"          - string\n" +
			"        type: protein\n" +
			"        value: value\n" +
			"        cyclic: true\n" +
			"        modifications:\n" +
			"          - residue_index: 0\n" +
			"            type: ccd\n" +
			"            value: value\n" +
			"    id: id\n" +
			"target:\n" +
			"  chain_selection:\n" +
			"    A:\n" +
			"      chain_type: polymer\n" +
			"      crop_residues:\n" +
			"        - 0\n" +
			"        - 1\n" +
			"        - 2\n" +
			"        - 3\n" +
			"        - 4\n" +
			"        - 5\n" +
			"        - 6\n" +
			"        - 7\n" +
			"        - 8\n" +
			"        - 9\n" +
			"        - 10\n" +
			"        - 11\n" +
			"        - 12\n" +
			"      epitope_residues:\n" +
			"        - 10\n" +
			"        - 11\n" +
			"        - 12\n" +
			"      flexible_residues:\n" +
			"        - 5\n" +
			"        - 6\n" +
			"        - 7\n" +
			"      non_binding_residues:\n" +
			"        - 0\n" +
			"        - 1\n" +
			"        - 2\n" +
			"  structure:\n" +
			"    type: url\n" +
			"    url: https://example.com\n" +
			"  type: structure_template\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"protein:library-screen", "start",
		)
	})
}

func TestProteinLibraryScreenStop(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:library-screen", "stop",
			"--id", "id",
		)
	})
}
