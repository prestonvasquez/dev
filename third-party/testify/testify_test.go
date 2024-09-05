package main

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertErrorIs(t *testing.T) {
	assert.ErrorIs(t, io.EOF, errors.New("custom error"))
}
