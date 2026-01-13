package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/pipekit/databricks-connector/pkg/common"
	"github.com/spf13/cobra"
)

var (
	getOutputRunID   int64
	outputJSONPath   string
	outputURLPath    string
	outputResultPath string
	outputStatePath  string
)

var getOutputCmd = &cobra.Command{
	Use:   "get-output",
	Short: "Get output of a Databricks job run",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		w, err := common.GetDatabricksClient(ctx)
		if err != nil {
			return err
		}

		// Get Run Details (for URL and State)
		run, err := w.Jobs.GetRun(ctx, jobs.GetRunRequest{RunId: getOutputRunID})
		if err != nil {
			return fmt.Errorf("failed to get run: %w", err)
		}

		if outputURLPath != "" {
			if err := os.WriteFile(outputURLPath, []byte(run.RunPageUrl), 0644); err != nil {
				return fmt.Errorf("failed to write URL to file: %w", err)
			}
		}

		if outputStatePath != "" {
			state := ""
			if run.State != nil {
				state = string(run.State.LifeCycleState)
				if run.State.ResultState != "" {
					state += "/" + string(run.State.ResultState)
				}
			}
			if err := os.WriteFile(outputStatePath, []byte(state), 0644); err != nil {
				return fmt.Errorf("failed to write state to file: %w", err)
			}
		}

		// Get Run Output (for Notebook results or logs)
		// Note: GetRunOutput usually works for Notebook tasks.
		// For Spark tasks, it might return driver logs if available.
		output, err := w.Jobs.GetRunOutput(ctx, jobs.GetRunOutputRequest{RunId: getOutputRunID})
		if err == nil {
			// If we successfully got output
			if outputResultPath != "" {
				// Notebook output usually has 'Logs' or 'NotebookOutput'
				var content string
				if output.NotebookOutput != nil {
					// result might be in specific field
					if output.NotebookOutput.Result != "" {
						content = output.NotebookOutput.Result
					}
				} else if output.Logs != "" {
					content = output.Logs
				} else if output.Error != "" {
					content = output.Error
				}

				if content != "" {
					if err := os.WriteFile(outputResultPath, []byte(content), 0644); err != nil {
						return fmt.Errorf("failed to write result to file: %w", err)
					}
				}
			}
			
			if outputJSONPath != "" {
				data, _ := json.MarshalIndent(output, "", "  ")
				if err := os.WriteFile(outputJSONPath, data, 0644); err != nil {
					return fmt.Errorf("failed to write output JSON: %w", err)
				}
			}
		} else {
			// It might not be a failure if the task type doesn't support GetRunOutput in the same way,
			// or if logs aren't available yet. But usually for terminated runs it is.
			fmt.Printf("Warning: Could not fetch run output: %v\n", err)
		}
		
		// Print URL to stdout as well just in case
		fmt.Printf("Run URL: %s\n", run.RunPageUrl)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getOutputCmd)
	getOutputCmd.Flags().Int64Var(&getOutputRunID, "run-id", 0, "ID of the run")
	getOutputCmd.Flags().StringVar(&outputJSONPath, "write-json", "", "Path to write full output JSON")
	getOutputCmd.Flags().StringVar(&outputURLPath, "write-url", "", "Path to write Run Page URL")
	getOutputCmd.Flags().StringVar(&outputResultPath, "write-result", "", "Path to write result value (e.g. notebook exit value)")
	getOutputCmd.Flags().StringVar(&outputStatePath, "write-state", "", "Path to write run state")
	getOutputCmd.MarkFlagRequired("run-id")
}
