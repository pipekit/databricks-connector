import sys
import random
from pyspark.sql import SparkSession

def main():
    spark = SparkSession.builder.appName("SimpleJob").getOrCreate()
    
    # Simple argument parsing
    # Args are passed as: file.py arg1 arg2 ...
    name = "World"
    if len(sys.argv) > 1:
        # sys.argv[0] is the script name
        # If passed via --parameters in our CLI, we need to see how they arrive.
        # Our CLI passes them as ["--key", "value"].
        # So we might need to parse flags or just look at positional args if we changed the CLI logic.
        # Wait, our CLI logic for spark-python was:
        # list = append(list, fmt.Sprintf("--%s", k), v)
        # So args will be: ["--name", "Alice"]
        try:
            name_idx = sys.argv.index("--name")
            if name_idx + 1 < len(sys.argv):
                name = sys.argv[name_idx + 1]
        except ValueError:
            pass

    print(f"Hello from Spark Job, {name}!")

    # create a simple DataFrame
    data = [("Alice", 1), ("Bob", 2), ("Charlie", 3)]
    df = spark.createDataFrame(data, ["Name", "Value"])
    df.show()

    print("Job finished successfully")

if __name__ == "__main__":
    main()
