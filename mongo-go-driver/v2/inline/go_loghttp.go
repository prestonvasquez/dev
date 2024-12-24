package main

import (
	"context"
	"net/http"

	"github.com/motemen/go-loghttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type nonDefaultTransport struct{}

var _ http.RoundTripper = &nonDefaultTransport{}

func (*nonDefaultTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, nil }

func main() {
	opts := options.Client().SetHTTPClient(&http.Client{Transport: loghttp.DefaultTransport})

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()
}
