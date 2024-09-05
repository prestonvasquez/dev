//go:build ClientEncryption

package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	client, _ := mongo.Connect()
	client.Disconnect(context.Background())

	kmsProviders := map[string]map[string]interface{}{
		"local": {
			"key": make([]byte, 96),
		},
	}

	opts := options.ClientEncryption().SetKmsProviders(kmsProviders).SetKeyVaultNamespace("db.coll")
	ce, err := mongo.NewClientEncryption(client, opts)
	if err != nil {
		panic(err)
	}

	db := client.Database("db")

	var encryptedFields bson.Raw
	err = bson.UnmarshalExtJSON([]byte(`{
				"fields": [{
					"path": "ssn",
					"bsonType": "string",
					"keyId": null
				}]
			}`), true /* canonical */, &encryptedFields)

	cecOpts := options.CreateCollection().SetEncryptedFields(encryptedFields)
	_, _, err = ce.CreateEncryptedCollection(context.Background(), db, "coll", cecOpts, "local", nil)
	if err != nil {
		panic(err)
	}
}
