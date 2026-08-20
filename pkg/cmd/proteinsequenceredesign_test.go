// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
)

func TestProteinSequenceRedesignRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "retrieve",
			"--id", "id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinSequenceRedesignList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestProteinSequenceRedesignDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "delete-data",
			"--id", "id",
		)
	})
}

func TestProteinSequenceRedesignEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "estimate-cost",
			"--entity", "{chain_id: x, role: target, type: from_template}",
			"--entity", "{chain_id: x, role: target, type: from_template}",
			"--num-proteins", "1",
			"--structure", "{type: url, url: https://example.com}",
			"--type", "binder",
			"--global-design-filter", "{amino_acids: [I], type: excluded_amino_acids}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"entities:\n" +
			"  - chain_id: x\n" +
			"    role: target\n" +
			"    type: from_template\n" +
			"  - chain_id: x\n" +
			"    role: target\n" +
			"    type: from_template\n" +
			"num_proteins: 1\n" +
			"structure:\n" +
			"  type: url\n" +
			"  url: https://example.com\n" +
			"type: binder\n" +
			"global_design_filters:\n" +
			"  - amino_acids:\n" +
			"      - I\n" +
			"    type: excluded_amino_acids\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"protein:sequence-redesign", "estimate-cost",
		)
	})
}

func TestProteinSequenceRedesignListResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "list-results",
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

func TestProteinSequenceRedesignResume(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "resume",
			"--id", "id",
		)
	})
}

func TestProteinSequenceRedesignStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "start",
			"--entity", "{chain_id: x, role: target, type: from_template}",
			"--entity", "{chain_id: x, role: target, type: from_template}",
			"--num-proteins", "1",
			"--structure", "{type: url, url: https://example.com}",
			"--type", "binder",
			"--global-design-filter", "{amino_acids: [I], type: excluded_amino_acids}",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"entities:\n" +
			"  - chain_id: x\n" +
			"    role: target\n" +
			"    type: from_template\n" +
			"  - chain_id: x\n" +
			"    role: target\n" +
			"    type: from_template\n" +
			"num_proteins: 1\n" +
			"structure:\n" +
			"  type: url\n" +
			"  url: https://example.com\n" +
			"type: binder\n" +
			"global_design_filters:\n" +
			"  - amino_acids:\n" +
			"      - I\n" +
			"    type: excluded_amino_acids\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"protein:sequence-redesign", "start",
		)
	})
}

func TestProteinSequenceRedesignStop(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"protein:sequence-redesign", "stop",
			"--id", "id",
		)
	})
}
