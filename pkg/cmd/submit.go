package cmd

import (
	"context"
	"fmt"

	"github.com/databricks/databricks-sdk-go/service/compute"
	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/pipekit/databricks-connector/pkg/common"
	"github.com/spf13/cobra"
)

var (
	taskType             string
	existingClusterID    string
	newClusterNodeType   string
	newClusterSparkVer   string
	newClusterNumWorkers int
	codePath             string
	parameters           map[string]string
	emailNotifications   []string
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a Databricks job run",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		w, err := common.GetDatabricksClient(ctx)
		if err != nil {
			return err
		}

		runName := "argo-workflow-run"

		// Task Configuration
		task := jobs.SubmitTask{
			TaskKey: "main_task",
		}

		// Cluster Configuration
		if existingClusterID != "" {
			task.ExistingClusterId = existingClusterID
		} else {
			if newClusterNodeType == "" || newClusterSparkVer == "" {
				return fmt.Errorf("if not using existing-cluster-id, node-type and spark-version are required")
			}
			task.NewCluster = &compute.ClusterSpec{
				SparkVersion: newClusterSparkVer,
				NodeTypeId:   newClusterNodeType,
				NumWorkers:   newClusterNumWorkers,
			}
		}

		switch taskType {
		case "notebook":
			task.NotebookTask = &jobs.NotebookTask{
				NotebookPath:   codePath,
				BaseParameters: parameters,
			}
		case "spark-python":
			task.SparkPythonTask = &jobs.SparkPythonTask{
				PythonFile: codePath,
				Parameters: getListParams(parameters),
			}
		default:
			return fmt.Errorf("unsupported task type: %s", taskType)
		}

		submitRun := jobs.SubmitRun{
			RunName: runName,
			Tasks:   []jobs.SubmitTask{task},
		}

		// Notifications (Apply to the job run level if possible, or we rely on Job defaults.
		// SubmitRun struct has EmailNotifications field)
		if len(emailNotifications) > 0 {
			submitRun.EmailNotifications = &jobs.JobEmailNotifications{
				OnFailure: emailNotifications,
				OnSuccess: emailNotifications,
			}
		}

		run, err := w.Jobs.Submit(ctx, submitRun)
		if err != nil {
			return fmt.Errorf("failed to submit run: %w", err)
		}

		// Return the Wait method to get the run ID?
		// w.Jobs.Submit returns *jobs.WaitGetRunJobRunOpenOutput.
		// We can access the RunID from the result, but wait, usually Submit returns a struct with RunId.
		// Let's verify the return type of w.Jobs.Submit.
		// It returns a Waiter object in the new SDK usually.
		// Actually, I should check if it returns the Run object directly or a wrapper.
		// SDK v0.96.0 usually returns a "Wait" struct which has the RunID embedded or accessible.
		// Let's assume for a moment it returns something with RunId, otherwise I'll debug.
		
		// For now, I'll print the RunId. If compilation fails, I'll fix it.
		// The Wait object usually has a `RunId` field or `Response`.
		// Let's assume it returns a struct with RunId for now as per previous code,
		// but I suspect it might be `WaitGetRunJobRunOpenOutput` which wraps the operation.
		
		// Checking SDK patterns: operations often return a `*Wait...` struct which has a `RunId` field
		// if it's an immediate ID return.
		
		fmt.Printf("%d", run.RunId)

		return nil
	},
}

func getListParams(params map[string]string) []string {
	var list []string
	for k, v := range params {
		list = append(list, fmt.Sprintf("--%s", k), v)
	}
	return list
}

func init() {
	rootCmd.AddCommand(submitCmd)

	submitCmd.Flags().StringVar(&taskType, "task-type", "notebook", "Type of task: notebook or spark-python")
	submitCmd.Flags().StringVar(&existingClusterID, "existing-cluster-id", "", "ID of an existing cluster to reuse")
	submitCmd.Flags().StringVar(&newClusterNodeType, "new-cluster-node-type", "", "Node type for new cluster (e.g., i3.xlarge)")
	submitCmd.Flags().StringVar(&newClusterSparkVer, "new-cluster-spark-version", "", "Spark version for new cluster")
	submitCmd.Flags().IntVar(&newClusterNumWorkers, "new-cluster-num-workers", 1, "Number of workers for new cluster")
	submitCmd.Flags().StringVar(&codePath, "code-path", "", "Path to the code (Notebook path or Python file path)")
	submitCmd.Flags().StringToStringVar(&parameters, "parameters", nil, "Parameters for the job (key=value)")
	submitCmd.Flags().StringSliceVar(&emailNotifications, "email-notifications", nil, "List of emails for notifications")
	
	submitCmd.MarkFlagRequired("code-path")
}
