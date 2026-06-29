// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
)

func TestShareLinksCreate(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"share-links", "create",
			"--expires-at", "expires_at",
			"--pipeline-id", "string",
			"--prediction-id", "string",
			"--workspace-id", "workspace_id",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"expires_at: expires_at\n" +
			"pipeline_ids:\n" +
			"  - string\n" +
			"prediction_ids:\n" +
			"  - string\n" +
			"workspace_id: workspace_id\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"share-links", "create",
		)
	})
}

func TestShareLinksRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"share-links", "retrieve",
			"--id", "shr_qoEFr2BlPTBLuM5BinaC8x7iVPP_AwppEOmlxQjJ-eo",
		)
	})
}

func TestShareLinksDeleteData(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"share-links", "delete-data",
			"--id", "shr_qoEFr2BlPTBLuM5BinaC8x7iVPP_AwppEOmlxQjJ-eo",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("id: shr_qoEFr2BlPTBLuM5BinaC8x7iVPP_AwppEOmlxQjJ-eo")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"share-links", "delete-data",
		)
	})
}

func TestShareLinksListPipelineResults(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"share-links", "list-pipeline-results",
			"--max-items", "10",
			"--id", "id",
			"--pipeline-id", "pipelineId",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--ids", "ids",
			"--limit", "1",
		)
	})
}
