// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestSmallMoleculeDesignRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "retrieve",
			"--id", "id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeDesignList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeDesignDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "delete-data",
			"--id", "id",
		)
	})
}

func TestSmallMoleculeDesignEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "estimate-cost",
			"--num-molecules", "10",
			"--target", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], bonds: [{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}], constraints: [{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}], pocket_residues: {A: [42, 43, 44, 67, 68, 69]}, precomputed_affinity_pocket: {method: reference_ligands, pocket_file_url: s3://J!/&*z57, structure_file_url: s3://J!/&*z57}, reference_ligands: [string], type: no_template}",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters", "{boltz_smarts_catalog_filter_level: recommended, custom_filters: [{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]}",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(smallMoleculeDesignEstimateCost)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "estimate-cost",
			"--num-molecules", "10",
			"--target.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--target.bonds", "[{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}]",
			"--target.constraints", "[{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}]",
			"--target.pocket-residues", "{A: [42, 43, 44, 67, 68, 69]}",
			"--target.precomputed-affinity-pocket", "{method: reference_ligands, pocket_file_url: s3://J!/&*z57, structure_file_url: s3://J!/&*z57}",
			"--target.reference-ligands", "[string]",
			"--target.type", "no_template",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters.boltz-smarts-catalog-filter-level", "recommended",
			"--molecule-filters.custom-filters", "[{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"num_molecules: 10\n" +
			"target:\n" +
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
			"  pocket_residues:\n" +
			"    A:\n" +
			"      - 42\n" +
			"      - 43\n" +
			"      - 44\n" +
			"      - 67\n" +
			"      - 68\n" +
			"      - 69\n" +
			"  precomputed_affinity_pocket:\n" +
			"    method: reference_ligands\n" +
			"    pocket_file_url: s3://J!/&*z57\n" +
			"    structure_file_url: s3://J!/&*z57\n" +
			"  reference_ligands:\n" +
			"    - string\n" +
			"  type: no_template\n" +
			"chemical_space: enamine_real\n" +
			"idempotency_key: idempotency_key\n" +
			"molecule_filters:\n" +
			"  boltz_smarts_catalog_filter_level: recommended\n" +
			"  custom_filters:\n" +
			"    - max_hba: 0\n" +
			"      max_hbd: 0\n" +
			"      max_logp: 0\n" +
			"      max_mw: 0\n" +
			"      type: lipinski_filter\n" +
			"      allow_single_violation: true\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"small-molecule:design", "estimate-cost",
		)
	})
}

func TestSmallMoleculeDesignListResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "list-results",
			"--max-items", "10",
			"--id", "id",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeDesignStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "start",
			"--num-molecules", "10",
			"--target", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], bonds: [{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}], constraints: [{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}], pocket_residues: {A: [42, 43, 44, 67, 68, 69]}, precomputed_affinity_pocket: {method: reference_ligands, pocket_file_url: s3://J!/&*z57, structure_file_url: s3://J!/&*z57}, reference_ligands: [string], type: no_template}",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters", "{boltz_smarts_catalog_filter_level: recommended, custom_filters: [{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]}",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(smallMoleculeDesignStart)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "start",
			"--num-molecules", "10",
			"--target.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--target.bonds", "[{atom1: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, type: ligand_atom}}]",
			"--target.constraints", "[{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}]",
			"--target.pocket-residues", "{A: [42, 43, 44, 67, 68, 69]}",
			"--target.precomputed-affinity-pocket", "{method: reference_ligands, pocket_file_url: s3://J!/&*z57, structure_file_url: s3://J!/&*z57}",
			"--target.reference-ligands", "[string]",
			"--target.type", "no_template",
			"--chemical-space", "enamine_real",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters.boltz-smarts-catalog-filter-level", "recommended",
			"--molecule-filters.custom-filters", "[{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"num_molecules: 10\n" +
			"target:\n" +
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
			"  pocket_residues:\n" +
			"    A:\n" +
			"      - 42\n" +
			"      - 43\n" +
			"      - 44\n" +
			"      - 67\n" +
			"      - 68\n" +
			"      - 69\n" +
			"  precomputed_affinity_pocket:\n" +
			"    method: reference_ligands\n" +
			"    pocket_file_url: s3://J!/&*z57\n" +
			"    structure_file_url: s3://J!/&*z57\n" +
			"  reference_ligands:\n" +
			"    - string\n" +
			"  type: no_template\n" +
			"chemical_space: enamine_real\n" +
			"idempotency_key: idempotency_key\n" +
			"molecule_filters:\n" +
			"  boltz_smarts_catalog_filter_level: recommended\n" +
			"  custom_filters:\n" +
			"    - max_hba: 0\n" +
			"      max_hbd: 0\n" +
			"      max_logp: 0\n" +
			"      max_mw: 0\n" +
			"      type: lipinski_filter\n" +
			"      allow_single_violation: true\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"small-molecule:design", "start",
		)
	})
}

func TestSmallMoleculeDesignStop(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:design", "stop",
			"--id", "id",
		)
	})
}
