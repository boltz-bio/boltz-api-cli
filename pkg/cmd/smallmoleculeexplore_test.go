// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestSmallMoleculeExploreRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:explore", "retrieve",
			"--id", "id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestSmallMoleculeExploreListResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:explore", "list-results",
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

func TestSmallMoleculeExploreResume(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:explore", "resume",
			"--id", "id",
		)
	})
}

func TestSmallMoleculeExploreStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:explore", "start",
			"--budget", "1",
			"--library", "{format: csv, source: {type: url, url: https://example.com}, id_column: id_column, smiles_column: smiles_column}",
			"--target", "{entities: [{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}], bonds: [{atom1: {atom_name: atom_name, chain_id: chain_id, residue_index: 0, type: polymer_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, residue_index: 0, type: polymer_atom}}], constraints: [{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}], pocket_residues: {A: [42, 43, 44, 67, 68, 69]}, reference_ligands: [string], type: no_template}",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters", "{boltz_smarts_catalog_filter_level: recommended, custom_filters: [{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]}",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(smallMoleculeExploreStart)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:explore", "start",
			"--budget", "1",
			"--library.format", "csv",
			"--library.source", "{type: url, url: https://example.com}",
			"--library.id-column", "id_column",
			"--library.smiles-column", "smiles_column",
			"--target.entities", "[{chain_ids: [string], type: protein, value: value, cyclic: true, modifications: [{residue_index: 0, type: ccd, value: value}]}]",
			"--target.bonds", "[{atom1: {atom_name: atom_name, chain_id: chain_id, residue_index: 0, type: polymer_atom}, atom2: {atom_name: atom_name, chain_id: chain_id, residue_index: 0, type: polymer_atom}}]",
			"--target.constraints", "[{binder_chain_id: binder_chain_id, contact_residues: {A: [42, 43, 44, 67, 68, 69]}, max_distance_angstrom: 0, type: pocket, force: true}]",
			"--target.pocket-residues", "{A: [42, 43, 44, 67, 68, 69]}",
			"--target.reference-ligands", "[string]",
			"--target.type", "no_template",
			"--idempotency-key", "idempotency_key",
			"--molecule-filters.boltz-smarts-catalog-filter-level", "recommended",
			"--molecule-filters.custom-filters", "[{max_hba: 0, max_hbd: 0, max_logp: 0, max_mw: 0, type: lipinski_filter, allow_single_violation: true}]",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"budget: 1\n" +
			"library:\n" +
			"  format: csv\n" +
			"  source:\n" +
			"    type: url\n" +
			"    url: https://example.com\n" +
			"  id_column: id_column\n" +
			"  smiles_column: smiles_column\n" +
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
			"        residue_index: 0\n" +
			"        type: polymer_atom\n" +
			"      atom2:\n" +
			"        atom_name: atom_name\n" +
			"        chain_id: chain_id\n" +
			"        residue_index: 0\n" +
			"        type: polymer_atom\n" +
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
			"  reference_ligands:\n" +
			"    - string\n" +
			"  type: no_template\n" +
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
			"small-molecule:explore", "start",
		)
	})
}

func TestSmallMoleculeExploreStop(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"small-molecule:explore", "stop",
			"--id", "id",
		)
	})
}
