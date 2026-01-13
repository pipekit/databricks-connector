package common

import (
	"context"
	"fmt"

	"github.com/databricks/databricks-sdk-go"
)

// GetDatabricksClient initializes and returns a Databricks Workspace Client.
// It relies on the SDK's default configuration loading (env vars, .databrickscfg, etc.).
func GetDatabricksClient(ctx context.Context) (*databricks.WorkspaceClient, error) {
	w, err := databricks.NewWorkspaceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize databricks client: %w", err)
	}
	return w, nil
}
