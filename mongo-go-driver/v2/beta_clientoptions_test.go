package v2

import (
	"testing"
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

func TestBetaClientOptions(t *testing.T) {}
