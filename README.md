# Databricks Connector for Argo Workflows

This CLI tool is designed to be used as a step in Argo Workflows to submit, monitor, and retrieve results from Databricks jobs.

## Building

```bash
go build -o databricks-connector cmd/databricks-connector/main.go
```

## Usage

For a complete reference of all commands and flags, see the [CLI Reference](docs/databricks-connector.md).

### 1. Submit a Run

Submits a new run. Prints the Run ID to stdout.

**Notebook Task on New Cluster:**
```bash
./databricks-connector submit \
  --task-type notebook \
  --code-path /Users/me/my-notebook \
  --new-cluster-node-type i3.xlarge \
  --new-cluster-spark-version 13.3.x-scala2.12 \
  --new-cluster-num-workers 2 \
  --parameters "param1=value1,param2=value2"
```

**Spark Python Task on Existing Cluster:**
```bash
./databricks-connector submit \
  --task-type spark-python \
  --code-path dbfs:/FileStore/my-script.py \
  --existing-cluster-id 1234-567890-abcde \
  --parameters "arg1=val1"
```

### 2. Start an Existing Job

Triggers a run of an existing Databricks job.

```bash
./databricks-connector start \
  --job-id 123456 \
  --job-params "key1=value1"
```

### 3. Monitor a Run

Polls the run status and streams state changes. Blocks until completion.

```bash
./databricks-connector monitor --run-id <RUN_ID> --interval 10s
```

### 4. Get Outputs

Retrieves run details and outputs to files (for Argo Output Parameters).

```bash
./databricks-connector get-output \
  --run-id <RUN_ID> \
  --write-url /tmp/run_url.txt \
  --write-result /tmp/result.txt \
  --write-state /tmp/state.txt
```

### 5. Cancel a Run

Cancels an active run.

```bash
./databricks-connector cancel --run-id <RUN_ID>
```

## Kubernetes & Argo Setup

To use this connector within Argo Workflows, you need to deploy the Workflow Template and a Secret containing your Databricks credentials.

### 1. Configure Credentials

Edit `manifests/secret-example.yaml` with your Databricks Host URL and Token.

```bash
kubectl apply -f manifests/secret-example.yaml
```

### 2. Install Workflow Template

Apply the Workflow Template to your cluster. This template encapsulates the `submit` (or `start`), `monitor`, and `get-output` steps into a reusable `run-job` template.

```bash
kubectl apply -f manifests/workflow-template.yaml
```

### 3. Usage in a Workflow

You can now reference the `databricks-connector` template in your own workflows.

**Example: Submit a new Notebook run**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: databricks-run-
spec:
  entrypoint: main
  templates:
  - name: main
    steps:
    - - name: run-notebook
        templateRef:
          name: databricks-connector
          template: run-job
        arguments:
          parameters:
          - name: code-path
            value: "/Workspace/Users/me/my-notebook"
          - name: task-type
            value: "notebook"
          - name: cluster-mode
            value: "Existing"
          - name: existing-cluster-id
            value: "1234-567890-abcde"
```

**Example: Run an existing Job**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: databricks-job-run-
spec:
  entrypoint: main
  templates:
  - name: main
    steps:
    - - name: run-job
        templateRef:
          name: databricks-connector
          template: run-existing-job
        arguments:
          parameters:
          - name: job-id
            value: "987654"
```

## Examples

Check the `examples/` directory for ready-to-use Workflow manifests:

*   **`examples/spark-jar-workflow.yaml`**: Demonstrates running a Spark JAR task (includes a sample Java project).
*   **`examples/run-existing-job-workflow.yaml`**: Demonstrates triggering an existing Databricks Job by ID.
*   **`examples/my-databricks-project/`**: Contains sample Python scripts and notebooks for testing.

See `examples/README.md` for detailed build and usage instructions.
