package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/pipekit/databricks-connector/pkg/common"
	"github.com/spf13/cobra"
)

var (
	monitorRunID int64
	pollInterval time.Duration
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor a Databricks job run",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		w, err := common.GetDatabricksClient(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Monitoring run %d...\n", monitorRunID)

		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		var lastState jobs.RunLifeCycleState

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				run, err := w.Jobs.GetRun(ctx, jobs.GetRunRequest{RunId: monitorRunID})
				if err != nil {
					fmt.Printf("Error fetching run status: %v\n", err)
					continue
				}

				state := run.State
				if state == nil {
					continue
				}

				if state.LifeCycleState != lastState {
					fmt.Printf("Run State: %s (Result: %s) - %s\n", state.LifeCycleState, state.ResultState, state.StateMessage)
					lastState = state.LifeCycleState
				}

				if state.LifeCycleState == jobs.RunLifeCycleStateTerminated || 
				   state.LifeCycleState == jobs.RunLifeCycleStateSkipped || 
				   state.LifeCycleState == jobs.RunLifeCycleStateInternalError {
					
					if state.ResultState == jobs.RunResultStateSuccess {
						fmt.Println("Run completed successfully.")
						return nil
					} else {
						return fmt.Errorf("run failed with state: %s, result: %s, message: %s", state.LifeCycleState, state.ResultState, state.StateMessage)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().Int64Var(&monitorRunID, "run-id", 0, "ID of the run to monitor")
	monitorCmd.Flags().DurationVar(&pollInterval, "interval", 10*time.Second, "Polling interval")
	monitorCmd.MarkFlagRequired("run-id")
}
