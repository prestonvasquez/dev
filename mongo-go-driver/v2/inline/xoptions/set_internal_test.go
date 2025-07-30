package xoptions

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/xoptions"
)

func TestSetInternalInsertOneOptions(t *testing.T) {
	opts := options.InsertOne()
	err := xoptions.SetInternalInsertOneOptions(opts, "rawData", nil)
	fmt.Println(err)
}
