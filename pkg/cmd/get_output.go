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
	dumpJSON         bool
)

var getOutputCmd = &cobra.Command{
	Use:   "get-output",
	Short: "Get output of a Databricks job run",
	Long: `Retrieve the outputs, state, and details of a completed Databricks job run.

This command is designed to extract specific pieces of information (URL, State, Result)
and write them to files. This is particularly useful for Argo Workflows "Output Parameters",
where you can map the file content to a workflow variable.

It can also dump the full JSON representation of the run and its output for debugging.`,
	Example: `  # Write output details to specific files for Argo
  databricks-connector get-output \
    --run-id 123456 \
    --write-url /tmp/run_url \
    --write-result /tmp/result \
    --write-state /tmp/state

  # Dump full JSON to stdout
  databricks-connector get-output --run-id 123456 --json`,
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
		output, outputErr := w.Jobs.GetRunOutput(ctx, jobs.GetRunOutputRequest{RunId: getOutputRunID})
		
		taskOutputs := make(map[string]*jobs.RunOutput)
		var firstTaskOutput *jobs.RunOutput

		// If root output failed or we want to be thorough, check sub-tasks
		if len(run.Tasks) > 0 {
			for _, task := range run.Tasks {
				tOutput, err := w.Jobs.GetRunOutput(ctx, jobs.GetRunOutputRequest{RunId: task.RunId})
				if err == nil {
					taskOutputs[task.TaskKey] = tOutput
					if firstTaskOutput == nil {
						firstTaskOutput = tOutput
					}
				} else {
					fmt.Printf("Warning: Could not fetch output for task %s (RunID: %d): %v\n", task.TaskKey, task.RunId, err)
				}
			}
		}

		// Decide which output to use for 'result' file and 'output' variable
		// If root output is valid (and not just empty), use it. 
		// Otherwise, fallback to first task output.
		finalOutput := output
		if outputErr != nil || (output != nil && output.NotebookOutput == nil && output.Logs == "" && output.Error == "") {
			if firstTaskOutput != nil {
				finalOutput = firstTaskOutput
				outputErr = nil // Clear error as we found a fallback
			}
		}

		if outputErr == nil && finalOutput != nil {
			// If we successfully got output
			if outputResultPath != "" {
				// Notebook output usually has 'Logs' or 'NotebookOutput'
				var content string
				if finalOutput.NotebookOutput != nil {
					// result might be in specific field
					if finalOutput.NotebookOutput.Result != "" {
						content = finalOutput.NotebookOutput.Result
					}
				} else if finalOutput.Logs != "" {
					content = finalOutput.Logs
				} else if finalOutput.Error != "" {
					content = finalOutput.Error
				}

				if content != "" {
					if err := os.WriteFile(outputResultPath, []byte(content), 0644); err != nil {
						return fmt.Errorf("failed to write result to file: %w", err)
					}
				}
			}
		} else {
			// It might not be a failure if the task type doesn't support GetRunOutput in the same way,
			// or if logs aren't available yet. But usually for terminated runs it is.
			if !dumpJSON {
				fmt.Printf("Warning: Could not fetch run output: %v\n", outputErr)
			}
		}

		// Always write JSON if requested, combining Run and Output (and Error if any)
		if outputJSONPath != "" {
			combined := struct {
				Run         *jobs.Run                  `json:"run"`
				Output      *jobs.RunOutput            `json:"output,omitempty"`
				TaskOutputs map[string]*jobs.RunOutput `json:"task_outputs,omitempty"`
				Error       string                     `json:"error,omitempty"`
			}{
				Run:         run,
				Output:      finalOutput,
				TaskOutputs: taskOutputs,
			}
			if outputErr != nil {
				combined.Error = outputErr.Error()
			}

			data, err := json.MarshalIndent(combined, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal output JSON: %w", err)
			}
			if err := os.WriteFile(outputJSONPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write output JSON: %w", err)
			}
		}

		if dumpJSON {
			combined := struct {
				Run         *jobs.Run                  `json:"run"`
				Output      *jobs.RunOutput            `json:"output,omitempty"`
				TaskOutputs map[string]*jobs.RunOutput `json:"task_outputs,omitempty"`
				Error       string                     `json:"error,omitempty"`
			}{
				Run:         run,
				Output:      finalOutput,
				TaskOutputs: taskOutputs,
			}
			if outputErr != nil {
				combined.Error = outputErr.Error()
			}
			data, err := json.MarshalIndent(combined, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Print URL to stdout as well just in case
		fmt.Printf("Run URL: %s\n", run.RunPageUrl)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getOutputCmd)
	getOutputCmd.Flags().Int64Var(&getOutputRunID, "run-id", 0, "ID of the run")
	getOutputCmd.Flags().StringVar(&outputJSONPath, "write-json", "", "Path to write full output JSON file")
	getOutputCmd.Flags().StringVar(&outputURLPath, "write-url", "", "Path to write Run Page URL")
	getOutputCmd.Flags().StringVar(&outputResultPath, "write-result", "", "Path to write result value (e.g. notebook exit value)")
	getOutputCmd.Flags().StringVar(&outputStatePath, "write-state", "", "Path to write run state")
	getOutputCmd.Flags().BoolVar(&dumpJSON, "json", false, "Dump full JSON of run and output to stdout")
	getOutputCmd.MarkFlagRequired("run-id")
}
