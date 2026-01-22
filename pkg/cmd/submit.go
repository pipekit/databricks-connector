package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	parametersList       []string
	emailNotifications   []string
	runName              string
	sparkConf            []string
	// New flags
	cloudProvider  string
	availability   string
	scalingType    string
	minWorkers     int
	maxWorkers     int
	clusterMode    string
	mainClassName  string
	argsList       []string
	dryRun         bool
	instancePoolID string
	policyID       string
	taskKey        string
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

		// Task Configuration

		task := jobs.SubmitTask{

			TaskKey: taskKey,
		}

		// Parse parametersList into parameters map

		parameters = make(map[string]string)

		for _, p := range parametersList {

			if p == "" {

				continue

			}

			parts := strings.SplitN(p, "=", 2)

			if len(parts) != 2 {

				return fmt.Errorf("invalid parameter format: %s (expected key=value)", p)

			}

			parameters[parts[0]] = parts[1]

		}

		if err := configureCluster(&task); err != nil {

			return err

		}

		if err := configureTaskDetails(&task); err != nil {
			return err
		}

		submitRun := jobs.SubmitRun{
			RunName: runName,
			Tasks:   []jobs.SubmitTask{task},
		}

		// Serverless Spark Python AND JAR tasks require an environment definition
		if clusterMode == "Serverless" {
			configureServerlessEnv(&submitRun)
		}

		configureNotifications(&submitRun)

		if dryRun {
			b, err := json.MarshalIndent(submitRun, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}
			fmt.Println(string(b))
			return nil
		}

		run, err := w.Jobs.Submit(ctx, submitRun)
		if err != nil {
			return fmt.Errorf("failed to submit run: %w", err)
		}

		fmt.Printf("%d", run.RunId)

		return nil
	},
}

func configureCluster(task *jobs.SubmitTask) error {
	switch clusterMode {
	case "Existing":
		if existingClusterID == "" {
			return fmt.Errorf("existing-cluster-id is required when cluster-mode is Existing")
		}
		task.ExistingClusterId = existingClusterID

	case "Serverless":
		// For Serverless, we intentionally do NOT set ExistingClusterId or NewCluster.
		// Databricks will use the default serverless compute for the workspace.
		fmt.Fprintln(os.Stderr, "Submitting job using Serverless compute...")

	case "New":
		// Parse sparkConf slice into map
		sparkConfMap := make(map[string]string)
		for _, s := range sparkConf {
			if s == "" {
				continue
			}
			parts := strings.SplitN(s, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid spark-conf format: %s (expected key=value)", s)
			}
			sparkConfMap[parts[0]] = parts[1]
		}

		// Only process new cluster params if mode is New
		clusterSpec := &compute.ClusterSpec{
			SparkVersion:   newClusterSparkVer,
			NodeTypeId:     newClusterNodeType,
			SparkConf:      sparkConfMap,
			InstancePoolId: instancePoolID,
			PolicyId:       policyID,
		}

		// Autoscaling logic
		if scalingType == "autoscale" {
			if minWorkers < 1 || maxWorkers < 1 {
				return fmt.Errorf("min-workers and max-workers must be > 0 for autoscale")
			}
			clusterSpec.Autoscale = &compute.AutoScale{
				MinWorkers: minWorkers,
				MaxWorkers: maxWorkers,
			}
		} else {
			// Fixed size
			clusterSpec.NumWorkers = newClusterNumWorkers
		}

		// Availability / Cloud Attributes logic
		if availability != "" {
			switch cloudProvider {
			case "AWS":
				var awsAvail compute.AwsAvailability
				if availability == "SPOT" {
					awsAvail = compute.AwsAvailabilitySpot
				} else {
					awsAvail = compute.AwsAvailabilityOnDemand
				}
				clusterSpec.AwsAttributes = &compute.AwsAttributes{
					Availability: awsAvail,
				}
			case "AZURE":
				var azureAvail compute.AzureAvailability
				if availability == "SPOT" {
					azureAvail = compute.AzureAvailabilitySpotAzure
				} else {
					azureAvail = compute.AzureAvailabilityOnDemandAzure
				}
				clusterSpec.AzureAttributes = &compute.AzureAttributes{
					Availability: azureAvail,
				}
			case "GCP":
				// GCP attributes handling if needed
			default:
				if cloudProvider != "" {
					return fmt.Errorf("unsupported cloud provider: %s", cloudProvider)
				}
			}
		}
		task.NewCluster = clusterSpec

	default:
		return fmt.Errorf("cluster-mode must be 'Existing', 'New', or 'Serverless'")
	}
	return nil
}

func configureTaskDetails(task *jobs.SubmitTask) error {
	// Determine parameters: prioritize argsList (positional) over parameters (key-value flags)
	var taskParams []string
	if len(argsList) > 0 {
		taskParams = argsList
	} else {
		taskParams = getListParams(parameters)
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
			Parameters: taskParams,
			Source:     jobs.SourceWorkspace,
		}
	case "spark-jar":
		if mainClassName == "" {
			return fmt.Errorf("main-class-name is required for spark-jar task")
		}
		task.SparkJarTask = &jobs.SparkJarTask{
			MainClassName: mainClassName,
			Parameters:    taskParams,
		}
		// For Classic/New clusters, we add the jar to libraries.
		if clusterMode != "Serverless" {
			task.Libraries = []compute.Library{
				{Jar: codePath},
			}
		}
	default:
		return fmt.Errorf("unsupported task type: %s", taskType)
	}
	return nil
}

func configureServerlessEnv(submitRun *jobs.SubmitRun) {
	envKey := "default_env"

	// We do not specify "Client" to let Databricks pick the default for the task type.
	// Explicitly setting "1" caused Python REPL errors in JAR tasks.
	envSpec := &compute.Environment{}

	if taskType == "spark-jar" {
		envSpec.JavaDependencies = []string{codePath}
	}

	submitRun.Environments = []jobs.JobEnvironment{
		{
			EnvironmentKey: envKey,
			Spec:           envSpec,
		},
	}

	// Link the task to this environment
	// We assume a single task job here as per current logic
	if len(submitRun.Tasks) > 0 {
		submitRun.Tasks[0].EnvironmentKey = envKey
	}
}

func configureNotifications(submitRun *jobs.SubmitRun) {
	if len(emailNotifications) > 0 {
		submitRun.EmailNotifications = &jobs.JobEmailNotifications{
			OnFailure: emailNotifications,
			OnSuccess: emailNotifications,
		}
	}
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
	submitCmd.Flags().StringSliceVar(&parametersList, "parameters", nil, "Parameters for the job (key=value)")
	submitCmd.Flags().StringVar(&runName, "run-name", "argo-workflow-run", "Name of the job run")
	submitCmd.Flags().StringSliceVar(&sparkConf, "spark-conf", nil, "Spark configuration (key=value)")
	submitCmd.Flags().StringSliceVar(&emailNotifications, "email-notifications", nil, "List of emails for notifications")

	// New flags
	submitCmd.Flags().StringVar(&cloudProvider, "cloud-provider", "", "Cloud provider (AWS, AZURE, GCP)")
	submitCmd.Flags().StringVar(&availability, "availability", "", "Availability (SPOT, ON_DEMAND)")
	submitCmd.Flags().StringVar(&scalingType, "scaling-type", "fixed", "Scaling type (fixed, autoscale)")
	submitCmd.Flags().IntVar(&minWorkers, "min-workers", 1, "Min workers for autoscale")
	submitCmd.Flags().IntVar(&maxWorkers, "max-workers", 2, "Max workers for autoscale")
	submitCmd.Flags().StringVar(&clusterMode, "cluster-mode", "New", "Cluster mode (Existing, New, Serverless)")
	submitCmd.Flags().StringVar(&mainClassName, "main-class-name", "", "Main class name for spark-jar task")
	submitCmd.Flags().StringSliceVar(&argsList, "args", nil, "List of positional arguments for the job (overrides parameters)")
	submitCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the job submission request as JSON and exit without submitting")
	submitCmd.Flags().StringVar(&instancePoolID, "instance-pool-id", "", "Instance pool ID for new cluster")
	submitCmd.Flags().StringVar(&policyID, "policy-id", "", "Policy ID for new cluster")
	submitCmd.Flags().StringVar(&taskKey, "task-key", "main_task", "Task key for the job task")

	submitCmd.MarkFlagRequired("code-path")
}
