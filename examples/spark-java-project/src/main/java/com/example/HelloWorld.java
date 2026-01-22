package com.example;

import org.apache.spark.sql.Dataset;
import org.apache.spark.sql.Row;
import org.apache.spark.sql.SparkSession;

public class HelloWorld {
    public static void main(String[] args) {
        SparkSession spark = SparkSession
                .builder()
                .appName("Java Spark Hello World")
                .getOrCreate();

        System.out.println("Hello, Databricks!");

        // Create a simple Dataset
        Dataset<Row> df = spark.read().json(spark.createDataset(
                java.util.Arrays.asList("{\"name\":\"Alice\", " +
                        "\"age\": 25}", "{\"name\":\"Bob\", \"age\": 30}"),
                org.apache.spark.sql.Encoders.STRING()));

        df.show();
        
        // Print args if any
        if (args.length > 0) {
            System.out.println("Arguments provided:");
            for (String arg : args) {
                System.out.println(arg);
            }
        }

        spark.stop();
    }
}
