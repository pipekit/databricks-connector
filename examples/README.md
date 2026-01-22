# Examples

## 1. Spark Java Project

A simple "Hello World" Spark application to test the `spark-jar` task type.

### Build the JAR

Requirements: Maven and Java 8+.

```bash
cd spark-java-project
mvn package
```

This will create `target/spark-hello-world-1.0-SNAPSHOT.jar`.

### Upload to Databricks

You need to upload the JAR to a location your Databricks cluster can access (e.g., DBFS, S3, ADLS).

**Using Databricks CLI:**
```bash
databricks fs cp target/spark-hello-world-1.0-SNAPSHOT.jar dbfs:/FileStore/jars/spark-hello-world.jar
```

## 2. Run the Workflow

Ensure you have applied the `WorkflowTemplate` from the root `manifests/` directory.

```bash
kubectl apply -f ../manifests/workflow-template.yaml
```

Submit the example workflow:

```bash
argo submit examples/spark-jar-workflow.yaml \
  -p jar-path="dbfs:/FileStore/jars/spark-hello-world.jar" \
  -p databricks-secret="databricks-secret"
```

