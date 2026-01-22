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

### 2. Monitor a Run

Polls the run status and streams state changes. Blocks until completion.

```bash
./databricks-connector monitor --run-id <RUN_ID>
```

### 3. Get Outputs

Retrieves run details and outputs to files (for Argo Output Parameters).

```bash
./databricks-connector get-output \
  --run-id <RUN_ID> \
  --write-url /tmp/run_url.txt \
  --write-result /tmp/result.txt \
  --write-state /tmp/state.txt
```

## Argo Integration Example

You can wrap this binary in a container and use it in a Workflow Template.

1. **Submit Step:** Run `submit` and capture stdout as an output parameter `run-id`.
2. **Monitor Step:** Run `monitor` using the `run-id` input.
3. **Output Step:** Run `get-output` to write results to files, which Argo captures as output parameters.

```
