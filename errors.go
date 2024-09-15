package fmesh

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/cycle"
)

type ErrorHandlingStrategy int

const (
	StopOnFirstError ErrorHandlingStrategy = iota
	StopOnFirstPanic
	IgnoreAll
)

var (
	ErrHitAnError                       = errors.New("f-mesh hit an error and will be stopped")
	ErrHitAPanic                        = errors.New("f-mesh hit a panic and will be stopped")
	ErrUnsupportedErrorHandlingStrategy = errors.New("unsupported error handling strategy")
)

func newFMeshStopError(err error, cycleResult *cycle.Result) error {
	return fmt.Errorf("%w (cycle #%d activation results: %v)", err, cycleResult.CycleNumber(), cycleResult.ActivationResults())
}
