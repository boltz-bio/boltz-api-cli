// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
)

func TestProteinDesignRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "retrieve",
			"--id", "id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinDesignList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinDesignDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "delete-data",
			"--id", "id",
		)
	})
}

func TestProteinDesignEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "estimate-cost",
			"--binder-specification", "{chain_selection: {B: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9], design_motifs: [{design_length_range: {max: 8, min: 4}, end_index: 5, start_index: 0, type: replacement}]}}, modality: peptide, structure: {type: url, url: https://example.com}, type: structure_template, rules: {excluded_amino_acids: [x], excluded_sequence_motifs: [string], max_hydrophobic_fraction: 0}}",
			"--num-proteins", "10",
			"--target", "{chain_selection: {A: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], epitope_residues: [10, 11, 12], flexible_residues: [5, 6, 7], non_binding_residues: [0, 1, 2]}}, structure: {type: url, url: https://example.com}, type: structure_template}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"binder_specification:\n" +
			"  chain_selection:\n" +
			"    B:\n" +
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
			"      design_motifs:\n" +
			"        - design_length_range:\n" +
			"            max: 8\n" +
			"            min: 4\n" +
			"          end_index: 5\n" +
			"          start_index: 0\n" +
			"          type: replacement\n" +
			"  modality: peptide\n" +
			"  structure:\n" +
			"    type: url\n" +
			"    url: https://example.com\n" +
			"  type: structure_template\n" +
			"  rules:\n" +
			"    excluded_amino_acids:\n" +
			"      - x\n" +
			"    excluded_sequence_motifs:\n" +
			"      - string\n" +
			"    max_hydrophobic_fraction: 0\n" +
			"num_proteins: 10\n" +
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
			"protein:design", "estimate-cost",
		)
	})
}

func TestProteinDesignListResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "list-results",
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

func TestProteinDesignResume(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "resume",
			"--id", "id",
		)
	})
}

func TestProteinDesignStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "start",
			"--binder-specification", "{chain_selection: {B: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9], design_motifs: [{design_length_range: {max: 8, min: 4}, end_index: 5, start_index: 0, type: replacement}]}}, modality: peptide, structure: {type: url, url: https://example.com}, type: structure_template, rules: {excluded_amino_acids: [x], excluded_sequence_motifs: [string], max_hydrophobic_fraction: 0}}",
			"--num-proteins", "10",
			"--target", "{chain_selection: {A: {chain_type: polymer, crop_residues: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12], epitope_residues: [10, 11, 12], flexible_residues: [5, 6, 7], non_binding_residues: [0, 1, 2]}}, structure: {type: url, url: https://example.com}, type: structure_template}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"binder_specification:\n" +
			"  chain_selection:\n" +
			"    B:\n" +
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
			"      design_motifs:\n" +
			"        - design_length_range:\n" +
			"            max: 8\n" +
			"            min: 4\n" +
			"          end_index: 5\n" +
			"          start_index: 0\n" +
			"          type: replacement\n" +
			"  modality: peptide\n" +
			"  structure:\n" +
			"    type: url\n" +
			"    url: https://example.com\n" +
			"  type: structure_template\n" +
			"  rules:\n" +
			"    excluded_amino_acids:\n" +
			"      - x\n" +
			"    excluded_sequence_motifs:\n" +
			"      - string\n" +
			"    max_hydrophobic_fraction: 0\n" +
			"num_proteins: 10\n" +
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
			"protein:design", "start",
		)
	})
}

func TestProteinDesignStop(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:design", "stop",
			"--id", "id",
		)
	})
}
