package component

import (
	"errors"
	"fmt"
)

var (
	errWaitingForInputs     = errors.New("component is waiting for some inputs")
	errWaitingForInputsKeep = fmt.Errorf("%w: do not clear input ports", errWaitingForInputs)
)

// NewErrWaitForInputs returns respective error
func NewErrWaitForInputs(keepInputs bool) error {
	if keepInputs {
		return errWaitingForInputsKeep
	}
	return errWaitingForInputs
}
