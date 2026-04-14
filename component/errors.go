package component

import (
	"errors"
	"fmt"
)

// These are control-flow signals, not actual failures.
// They instruct the scheduler how to proceed with the current component.
//
// For now we only have two variants, so sentinel errors are sufficient.
// If more behaviors are introduced, consider switching to a typed error.
var (
	// ErrWaitingForInputs is returned when you want the component to wait for some inputs and skip (default `behavior`) all input signals received in the current activation cycle.
	ErrWaitingForInputs = errors.New("component is waiting for some inputs")

	// ErrWaitingForInputsKeep is returned when you want the component to wait for some inputs and keep all input signals received in the current activation cycle.
	ErrWaitingForInputsKeep = fmt.Errorf("%w: do not clear input ports", ErrWaitingForInputs)
)
