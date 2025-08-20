package testocontainersgo

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestMongoDBWithAtlasLocal(t *testing.T) {
	ctx := context.Background()

	const imageName = "mongodb/mongodb-atlas-local:latest"

	container, err := mongodb.Run(ctx, imageName,
		mongodb.WithUsername("admin"),
		mongodb.WithPassword("password"),
		mongodb.WithReplicaSet("meep"),
	)
	if err != nil {
		dumpLogs(t, ctx, container)
	}
	require.NoError(t, err, "failed to start MongoDB container")

	host, err := container.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	port, err := container.MappedPort(ctx, "27017")
	require.NoError(t, err, "failed to get mapped port")

	uri := fmt.Sprintf("mongodb://%s:%s/?directConnection=true", host, port.Port())

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err, "failed to connect to MongoDB")

	defer func() {
		client.Disconnect(ctx)
	}()

	require.Eventually(t, func() bool {
		err := client.Ping(ctx, nil)
		return err == nil
	}, 60*time.Second, 5*time.Second)
}

// dumpLogs will dump the logs of the MongoDB Atlas Local container to the
// integration test output.
func dumpLogs(t *testing.T, ctx context.Context, ctr testcontainers.Container) {
	t.Helper()

	r, err := ctr.Logs(ctx)
	require.NoError(t, err)

	bytes, err := io.ReadAll(r)
	t.Logf("MongoDB Atlas Local logs:\n%s", string(bytes))
}
