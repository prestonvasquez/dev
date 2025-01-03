package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Here is the proposed options pattern for the beta:

type BetaOptions struct {
	X, Y string
}

type BetaOptionsBuilder struct {
	Opts []func(*BetaOptions) error
}

func Beta() *BetaOptionsBuilder {
	return &BetaOptionsBuilder{}
}

// List returns a list of AggergateOptions setter functions.
func (bo *BetaOptionsBuilder) List() []func(*BetaOptions) error {
	return bo.Opts
}

func (bo *BetaOptionsBuilder) SetX(x string) func(*BetaOptions) error {
	return func(bo *BetaOptions) error {
		bo.X = x

		return nil
	}
}

func (bo *BetaOptionsBuilder) SetY(y string) func(*BetaOptions) error {
	return func(bo *BetaOptions) error {
		bo.Y = y

		return nil
	}
}

type BetaLister[T any] interface {
	List() []func(*T) error
}

func getOptions[T any](mopts BetaLister[T]) (*T, error) {
	opts := new(T)

	for _, setterFn := range mopts.List() {
		if setterFn == nil {
			continue
		}

		if err := setterFn(opts); err != nil {
			return nil, err
		}
	}

	return opts, nil
}

func TestBetaClientOptions(t *testing.T) {
	opts1 := &BetaOptions{X: "x"}
	opts2 := &BetaOptions{Y: "y"}

	opts := Beta()

	opts.Opts = append(opts.Opts, func(o *BetaOptions) error {
		*o = *opts1

		return nil
	})

	opts.Opts = append(opts.Opts, func(o *BetaOptions) error {
		*o = *opts2

		return nil
	})

	mergedOpts, _ := getOptions(opts)
	assert.Equal(t, "x", mergedOpts.X)
	assert.Equal(t, "y", mergedOpts.Y)

}


clientBldr, err := GetReplicaSetClientOptions(&clusterConfig, TestConnRsName, &logger)

var clientOpts options.ClientOptions
for _, set := range clientBldr {
	_ = set(&clientOpts)
}

assert.NotNil(t, clientOpts)
assert.Nil(t, err)
credentials := clientOpts.Auth

assert.Equal(t, TestConnUsername, credentials.Username)
