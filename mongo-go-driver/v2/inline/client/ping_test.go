package client_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestPing(t *testing.T) {
	//const uri = "mongodb://127.0.0.1:28017/?directConnection=true"
	poolMonitor := event.PoolMonitor{
		Event: func(evt *event.PoolEvent) {
			if evt.Type == "ConnectionCreated" {
				fmt.Println("Connection created:", evt.Address)
			}
		},
	}

	clientOpts := options.Client().SetPoolMonitor(&poolMonitor)
	_, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	//err = client.Ping(context.Background(), nil)
	//require.NoError(t, err)
}

//func TestInternalClientOptions(t *testing.T) {
//	opts := options.Client()
//	_ = xoptions.SetInternalInsertOneOptions(opts, "rawData", true)
//}
