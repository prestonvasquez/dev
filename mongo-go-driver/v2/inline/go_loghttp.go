package main

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/motemen/go-loghttp/global" // Just this line!
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type nonDefaultTransport struct{}

var _ http.RoundTripper = &nonDefaultTransport{}

func (*nonDefaultTransport) RoundTrip(*http.Request) (*http.Response, error) {
	fmt.Println("happens")
	return nil, nil
}

func main() {
	//opts := options.Client().SetHTTPClient(&http.Client{Transport: loghttp.DefaultTransport})
	http.DefaultTransport = &nonDefaultTransport{}

	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	client.Ping(context.Background(), nil)
}
