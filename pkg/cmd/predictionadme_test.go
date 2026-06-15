// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestPredictionsAdmeRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "retrieve",
			"--id", "id",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestPredictionsAdmeList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestPredictionsAdmeDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "delete-data",
			"--id", "id",
		)
	})
}

func TestPredictionsAdmeEstimateCost(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "estimate-cost",
			"--input", "{molecules: [{smiles: x, id: x}]}",
			"--model", "adme-v1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(predictionsAdmeEstimateCost)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "estimate-cost",
			"--input.molecules", "[{smiles: x, id: x}]",
			"--model", "adme-v1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"input:\n" +
			"  molecules:\n" +
			"    - smiles: x\n" +
			"      id: x\n" +
			"model: adme-v1\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"predictions:adme", "estimate-cost",
		)
	})
}

func TestPredictionsAdmeStart(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "start",
			"--input", "{molecules: [{smiles: x, id: x}]}",
			"--model", "adme-v1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(predictionsAdmeStart)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"predictions:adme", "start",
			"--input.molecules", "[{smiles: x, id: x}]",
			"--model", "adme-v1",
			"--idempotency-key", "idempotency_key",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"input:\n" +
			"  molecules:\n" +
			"    - smiles: x\n" +
			"      id: x\n" +
			"model: adme-v1\n" +
			"idempotency_key: idempotency_key\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"predictions:adme", "start",
		)
	})
}
