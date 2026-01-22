from pyspark.sql import SparkSession
import sys

def run_jar_main(main_class, args):
    spark = SparkSession.builder.getOrCreate()
    sc = spark.sparkContext
    
    # Access the JVM via Py4J
    jvm = sc._jvm
    
    # Split package and class name
    parts = main_class.split('.')
    
    # Navigate the JVM package structure
    java_obj = jvm
    for part in parts:
        java_obj = getattr(java_obj, part)
        
    # java_obj is now the Class. call main.
    # main takes String[]
    # We need to convert python list to Java Array if possible, or just pass nothing for hello world
    
    # Simple reflection invocation for Hello World (no args)
    # java_obj.main([]) 
    
    # Handling args is complex with py4j array conversion. 
    # For this example, we assume no args or simple string args.
    
    gateway = sc._gateway
    java_args = gateway.new_array(gateway.jvm.java.lang.String, len(args))
    for i, arg in enumerate(args):
        java_args[i] = arg
        
    print(f"Invoking {main_class}.main()...")
    java_obj.main(java_args)
    print("Invocation complete.")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: jar_runner.py <main_class> [args...]")
        sys.exit(1)
        
    main_class_name = sys.argv[1]
    job_args = sys.argv[2:]
    
    run_jar_main(main_class_name, job_args)
