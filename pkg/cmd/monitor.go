package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
		// Create a context that listens for system signals (SIGINT, SIGTERM)
		// This handles the "Stop" button in Argo Workflows.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)
		defer stop()

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
				// Use Stderr to ensure the log is seen even if stdout is buffered/cutoff
				fmt.Fprintf(os.Stderr, "\n[SIGNAL] Received termination signal (SIGTERM/SIGINT). Cancelling Databricks run %d...\n", monitorRunID)

				// Use a fresh context for cancellation since the parent ctx is done.
				// Short timeout (5s) to ensure we send the request before the pod is SIGKILLed (usually 30s grace).
				cancelCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, err := w.Jobs.CancelRun(cancelCtx, jobs.CancelRun{RunId: monitorRunID})
				if err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to cancel run: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "[SUCCESS] Run cancellation requested successfully.\n")
				}

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
