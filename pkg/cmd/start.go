package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/databricks/databricks-sdk-go/service/jobs"
	"github.com/pipekit/databricks-connector/pkg/common"
	"github.com/spf13/cobra"
)

var (
	jobID                  int64
	startNotebookParams    []string
	startPythonParams      []string
	startJarParams         []string
	startSparkSubmitParams []string
	startJobParams         []string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an existing Databricks job",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		w, err := common.GetDatabricksClient(ctx)
		if err != nil {
			return err
		}

		runNow := jobs.RunNow{
			JobId: jobID,
		}

		if len(startJobParams) > 0 {
			params := make(map[string]string)
			for _, p := range startJobParams {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid job parameter format: %s (expected key=value)", p)
				}
				params[parts[0]] = parts[1]
			}
			runNow.JobParameters = params
		}

		if len(startNotebookParams) > 0 {
			params := make(map[string]string)
			for _, p := range startNotebookParams {
				parts := strings.SplitN(p, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid notebook parameter format: %s (expected key=value)", p)
				}
				params[parts[0]] = parts[1]
			}
			runNow.NotebookParams = params
		}

		if len(startPythonParams) > 0 {
			runNow.PythonParams = startPythonParams
		}

		if len(startJarParams) > 0 {
			runNow.JarParams = startJarParams
		}

		if len(startSparkSubmitParams) > 0 {
			runNow.SparkSubmitParams = startSparkSubmitParams
		}

		run, err := w.Jobs.RunNow(ctx, runNow)
		if err != nil {
			return fmt.Errorf("failed to start job: %w", err)
		}

		fmt.Printf("%d", run.RunId)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().Int64Var(&jobID, "job-id", 0, "ID of the job to start")
	startCmd.Flags().StringSliceVar(&startJobParams, "job-params", nil, "Job parameters (key=value) - preferred for multi-task jobs")
	startCmd.Flags().StringSliceVar(&startNotebookParams, "notebook-params", nil, "Notebook parameters (key=value) - legacy")
	startCmd.Flags().StringSliceVar(&startPythonParams, "python-params", nil, "Python parameters (positional) - legacy")
	startCmd.Flags().StringSliceVar(&startJarParams, "jar-params", nil, "JAR parameters (positional) - legacy")
	startCmd.Flags().StringSliceVar(&startSparkSubmitParams, "spark-submit-params", nil, "Spark submit parameters (positional) - legacy")

	startCmd.MarkFlagRequired("job-id")
}
