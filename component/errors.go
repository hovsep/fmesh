package component

import (
	"errors"
	"fmt"
)

const (
	// KeepAllInputs means all input ports must hold their signals till the next activation cycle.
	KeepAllInputs = true
	// SkipAllInputs means all input ports must be cleared before the next activation cycle (default behavior).
	SkipAllInputs = false
)

var (
	errWaitingForInputs     = errors.New("component is waiting for some inputs")
	errWaitingForInputsKeep = fmt.Errorf("%w: do not clear input ports", errWaitingForInputs)
)

// NewErrWaitForInputs returns the respective error.
func NewErrWaitForInputs(keepInputs bool) error {
	switch keepInputs {
	case KeepAllInputs:
		return errWaitingForInputsKeep
	case SkipAllInputs:
		fallthrough
	default:
		return errWaitingForInputs
	}
}
