n = 100  # Replace 100 with your actual number of queries
latencies = []

for i in range(n):
    # Record the start time
    start_time = time.time()
    
    # Fetch documents from the collection
    results = collection.find(q)
     
    # Record the end time
    end_time = time.time()
    
    # Calculate latency (in seconds) and store it
    latency = end_time - start_time
    latencies.append(latency)
