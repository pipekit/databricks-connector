# Databricks notebook source
# COMMAND ----------

dbutils.widgets.text("name", "World", "Name Param")
name = dbutils.widgets.get("name")

# COMMAND ----------

print(f"Hello, {name}!")

# COMMAND ----------

import random
result_value = random.randint(1, 100)
print(f"Generated result: {result_value}")

# COMMAND ----------

dbutils.notebook.exit(str(result_value))
