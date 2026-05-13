// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package cmd

import (
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/mocktest"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
)

func TestAdminWorkspacesCreate(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "create",
			"--data-retention", "{unit: hours, value: 1}",
			"--name", "x",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(adminWorkspacesCreate)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "create",
			"--data-retention.unit", "hours",
			"--data-retention.value", "1",
			"--name", "x",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"data_retention:\n" +
			"  unit: hours\n" +
			"  value: 1\n" +
			"name: x\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"admin:workspaces", "create",
		)
	})
}

func TestAdminWorkspacesRetrieve(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "retrieve",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestAdminWorkspacesUpdate(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "update",
			"--workspace-id", "workspace_id",
			"--data-retention", "{unit: hours, value: 1}",
			"--name", "x",
		)
	})

	t.Run("inner flags", func(t *testing.T) {
		// Check that inner flags have been set up correctly
		requestflag.CheckInnerFlags(adminWorkspacesUpdate)

		// Alternative argument passing style using inner flags
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "update",
			"--workspace-id", "workspace_id",
			"--data-retention.unit", "hours",
			"--data-retention.value", "1",
			"--name", "x",
		)
	})

	t.Run("piping data", func(t *testing.T) {
		// Test piping YAML data over stdin
		pipeData := []byte("" +
			"data_retention:\n" +
			"  unit: hours\n" +
			"  value: 1\n" +
			"name: x\n")
		mocktest.TestRunMockTestWithPipeAndFlags(
			t, pipeData,
			"--api-key", "string",
			"admin:workspaces", "update",
			"--workspace-id", "workspace_id",
		)
	})
}

func TestAdminWorkspacesList(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "list",
			"--max-items", "10",
			"--after-id", "after_id",
			"--before-id", "before_id",
			"--limit", "1",
			"--name", "x",
		)
	})
}

func TestAdminWorkspacesArchive(t *testing.T) {
	t.Skip("Mock server tests are disabled")
	t.Run("regular flags", func(t *testing.T) {
		mocktest.TestRunMockTestWithFlags(
			t,
			"--api-key", "string",
			"admin:workspaces", "archive",
			"--workspace-id", "workspace_id",
		)
	})
}
