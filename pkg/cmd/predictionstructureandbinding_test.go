// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestPredictionsStructureAndBindingRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "retrieve",
			"--id", "sab_pred_2X7Ab9Cd3Ef6Gh1JkLmN",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestPredictionsStructureAndBindingList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestPredictionsStructureAndBindingDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "delete-data",
			"--id", "sab_pred_2X7Ab9Cd3Ef6Gh1JkLmN",
		)
	})
}

func TestPredictionsStructureAndBindingEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "estimate-cost",
			"--input", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], binding: {binder_chain_id: binder_chain_id, type: ligand_protein_binding}, bonds: [{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}], constraints: [{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}], model_options: {recycling_steps: 1, sampling_steps: 1, step_scale: 1.3}, num_samples: 1}",
			"--model", "boltz-2.1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(predictionsStructureAndBindingEstimateCost)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "estimate-cost",
			"--input.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--input.binding", "{binder_chain_id: binder_chain_id, type: ligand_protein_binding}",
			"--input.bonds", "[{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}]",
			"--input.constraints", "[{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}]",
			"--input.model-options", "{recycling_steps: 1, sampling_steps: 1, step_scale: 1.3}",
			"--input.num-samples", "1",
			"--model", "boltz-2.1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"input:\n" +
			"  entities:\n" +
			"    - chain_ids:\n" +
			"        - string\n" +
			"      type: protein\n" +
			"      value: value\n" +
			"      cyclic: true\n" +
			"      modifications:\n" +
			"        - residue_index: 0\n" +
			"          type: ccd\n" +
			"          value: value\n" +
			"  binding:\n" +
			"    binder_chain_id: binder_chain_id\n" +
			"    type: ligand_protein_binding\n" +
			"  bonds:\n" +
			"    - atom1:\n" +
			"        atom_name: atom_name\n" +
			"        chain_id: chain_id\n" +
			"        type: ligand_atom\n" +
			"      atom2:\n" +
			"        atom_name: atom_name\n" +
			"        chain_id: chain_id\n" +
			"        type: ligand_atom\n" +
			"  constraints:\n" +
			"    - binder_chain_id: binder_chain_id\n" +
			"      contact_residues:\n" +
			"        A:\n" +
			"          - 42\n" +
			"          - 43\n" +
			"          - 44\n" +
			"          - 67\n" +
			"          - 68\n" +
			"          - 69\n" +
			"      max_distance_angstrom: 0\n" +
			"      type: pocket\n" +
			"      force: true\n" +
			"  model_options:\n" +
			"    recycling_steps: 1\n" +
			"    sampling_steps: 1\n" +
			"    step_scale: 1.3\n" +
			"  num_samples: 1\n" +
			"model: boltz-2.1\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"predictions:structure-and-binding", "estimate-cost",
		)
	})
}

func TestPredictionsStructureAndBindingStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "start",
			"--input", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], binding: {binder_chain_id: binder_chain_id, type: ligand_protein_binding}, bonds: [{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}], constraints: [{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}], model_options: {recycling_steps: 1, sampling_steps: 1, step_scale: 1.3}, num_samples: 1}",
			"--model", "boltz-2.1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(predictionsStructureAndBindingStart)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:structure-and-binding", "start",
			"--input.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--input.binding", "{binder_chain_id: binder_chain_id, type: ligand_protein_binding}",
			"--input.bonds", "[{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}]",
			"--input.constraints", "[{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}]",
			"--input.model-options", "{recycling_steps: 1, sampling_steps: 1, step_scale: 1.3}",
			"--input.num-samples", "1",
			"--model", "boltz-2.1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"input:\n" +
			"  entities:\n" +
			"    - chain_ids:\n" +
			"        - string\n" +
			"      type: protein\n" +
			"      value: value\n" +
			"      cyclic: true\n" +
			"      modifications:\n" +
			"        - residue_index: 0\n" +
			"          type: ccd\n" +
			"          value: value\n" +
			"  binding:\n" +
			"    binder_chain_id: binder_chain_id\n" +
			"    type: ligand_protein_binding\n" +
			"  bonds:\n" +
			"    - atom1:\n" +
			"        atom_name: atom_name\n" +
			"        chain_id: chain_id\n" +
			"        type: ligand_atom\n" +
			"      atom2:\n" +
			"        atom_name: atom_name\n" +
			"        chain_id: chain_id\n" +
			"        type: ligand_atom\n" +
			"  constraints:\n" +
			"    - binder_chain_id: binder_chain_id\n" +
			"      contact_residues:\n" +
			"        A:\n" +
			"          - 42\n" +
			"          - 43\n" +
			"          - 44\n" +
			"          - 67\n" +
			"          - 68\n" +
			"          - 69\n" +
			"      max_distance_angstrom: 0\n" +
			"      type: pocket\n" +
			"      force: true\n" +
			"  model_options:\n" +
			"    recycling_steps: 1\n" +
			"    sampling_steps: 1\n" +
			"    step_scale: 1.3\n" +
			"  num_samples: 1\n" +
			"model: boltz-2.1\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"predictions:structure-and-binding", "start",
		)
	})
}
