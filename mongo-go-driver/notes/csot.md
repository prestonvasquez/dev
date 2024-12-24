## Failpoints 

- If a socket timeout is not applied, i.e. [this block|https://github.com/mongodb/mongo-go-driver/blob/cf0348c9a63f6e4086e2c62aa62e70a9d3b3a986/x/mongo/driver/topology/connection.go#L403-L406] is commented out, then mongod will block for the entire failpoint's blockTimeMS. Regardless of the value of `maxTimeMS` passed to the operation.
