package cmd

import (
	"context"
	"fmt"

	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/pipekit/databricks-connector/pkg/common"
	"github.com/spf13/cobra"
)

var cancelRunID int64

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a Databricks job run",
	Long: `Request the cancellation of a running Databricks job.

This sends a cancellation request to the Databricks API. It does not wait for the
cancellation to complete.`,
	Example: `  databricks-connector cancel --run-id 123456`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		w, err := common.GetDatabricksClient(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Cancelling run %d...\n", cancelRunID)
		_, err = w.Jobs.CancelRun(ctx, jobs.CancelRun{RunId: cancelRunID})
		if err != nil {
			return fmt.Errorf("failed to cancel run: %w", err)
		}

		fmt.Println("Run cancellation requested successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cancelCmd)
	cancelCmd.Flags().Int64Var(&cancelRunID, "run-id", 0, "ID of the run to cancel")
	cancelCmd.MarkFlagRequired("run-id")
}

