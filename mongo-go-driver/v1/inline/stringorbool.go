package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func main() {
	extJsonStr := `{"WriteBlocking": true}`
	type ExampleDoc struct {
		WriteBlocking StringOrBool2
	}

	var resultStr ExampleDoc
	err := bson.UnmarshalExtJSON([]byte(extJsonStr), false, &resultStr)
	if err != nil {
		panic(err)
	}
}

type StringOrBool2 struct {
	Value any
}

func (sb *StringOrBool2) UnmarshalBSON(data []byte) error {
	fmt.Println("unmarshaling bson!")

	var raw bson.RawValue
	if err := bson.Unmarshal(data, &raw); err != nil {
		fmt.Println("meep")
		return err
	}

	switch raw.Type {
	case bson.TypeString:
		strVal, ok := raw.StringValueOK()
		if !ok {
			return fmt.Errorf("invalid string value")
		}
		sb.Value = strVal

	case bson.TypeBoolean:
		boolVal, ok := raw.BooleanOK()
		if !ok {
			return fmt.Errorf("invalid boolean value")
		}
		sb.Value = boolVal

	default:
		return fmt.Errorf("invalid bson type: %s", raw.Type)
	}

	fmt.Println("hmm")

	return nil
}

func (sb *StringOrBool2) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	fmt.Println("unmarshaling bson value!")

	switch t {
	case bson.TypeString:
		//var strVal string
		//if err := bson.Unmarshal(data, &strVal); err != nil {
		//	fmt.Println("data", string(data))
		//	fmt.Println("err", err)
		//	return err
		//}
		values, err := bson.Raw(data).Values()
		fmt.Println(values, err)

		//sb.Value = strVal

	case bson.TypeBoolean:
		var boolVal bool
		if err := bson.UnmarshalValue(t, data, &boolVal); err != nil {
			fmt.Println("data", string(data))
			fmt.Println("err", err)
			return err
		}

		fmt.Println(boolVal)

		//values, err := bson.Raw(data).Values()
		//fmt.Println(values, err)
		sb.Value = boolVal
		//
		//value := bson.RawValue{
		//	Type:  t,
		//	Value: data,
		//}

		//if err := value.Unmarshal()

	default:
		fmt.Println(" default")
		return fmt.Errorf("invalid bson type: %s", t)
	}

	return nil
}
