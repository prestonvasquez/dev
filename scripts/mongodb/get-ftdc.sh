#!/bin/bash

# Check if the MONGODB_URI environment variable is set
if [ -z "$MONGODB_URI" ]; then
  echo "Error: The 'MONGODB_URI' environment variable is not set."
  echo "Please set it before running the script. See the guide:"
  echo "https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable"
  exit 1
fi
# Define variables for the database and collection
DB_NAME="testdb"                  # Replace with your database name
COLLECTION_NAME="largecollection" # Replace with your collection name
DELAY=1                           # Delay interval in seconds
# Run the MongoDB shell command with the specified collection statistics loop
mongosh "$MONGODB_URI" --eval "
  while (true) {
    const stats = db.getSiblingDB('$DB_NAME').$COLLECTION_NAME.stats();
    stats.time = new Date();
    print(JSON.stringify(stats));
    sleep(1000 * $DELAY);
  }
"
