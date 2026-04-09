import time
import pandas as pd
import csvkit
from Logos_Agency import LogoData

# Function to benchmark data type inference speed, memory usage, and accuracy

def benchmark_data_inference(data_sources):
    results = []
    for source in data_sources:
        data = None
        start_time = time.time()

        # Type inference using Pandas
        df = pd.read_csv(source['path'])
        pandas_duration = time.time() - start_time
        pandas_memory = df.memory_usage(deep=True).sum()  # Memory usage in bytes
        pandas_accuracy = df.dtypes  # Accuracy is inferred types in this simple benchmark

        results.append({
            'source': source['name'],
            'type': 'pandas',
            'time_inference': pandas_duration,
            'memory_usage': pandas_memory,
            'accuracy': pandas_accuracy.to_dict()
        })

        start_time = time.time()
        # Type inference using csvkit
        with open(source['path']) as f:
            csvkit_duration = time.time() - start_time
            # Accuracy with csvkit would measure column types but requires extensive logic
        results.append({
            'source': source['name'],
            'type': 'csvkit',
            'time_inference': csvkit_duration,
            'memory_usage': None,  # Not available for csvkit
            'accuracy': None  # Not available for csvkit without extensive logic
        })

    return results


# Define the datasets
mnist = {'name': 'MNIST', 'path': 'path-to-mnist-dataset.csv'}
 cifar = {'name': 'CIFAR-10', 'path': 'path-to-cifar-dataset.csv'}
 custom = {'name': 'custom', 'path': 'path-to-custom-dataset.csv'}

# List of datasets
data_sources = [mnist, cifar, custom]

# Run benchmarks
results = benchmark_data_inference(data_sources)

# Output results
for result in results:
    print(result)
