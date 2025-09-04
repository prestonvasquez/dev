package json_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericDecoding(t *testing.T) {
	toDecode := []byte(`{"csfle": true}`)

	type myStruct struct {
		CSFLE bool `json:"csfle"`
	}

	ms := myStruct{}
	if err := json.Unmarshal(toDecode, &ms); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	assert.Equal(t, true, ms.CSFLE)
}
