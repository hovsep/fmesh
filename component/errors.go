package component

import "errors"

var (
	//@TODO: provide wrapper methods so exact input can be specified within error
	ErrWaitingForInputResetInputs = errors.New("component is not ready (waiting for one or more inputs). All inputs will be reset")
	ErrWaitingForInputKeepInputs  = errors.New("component is not ready (waiting for one or more inputs). All inputs will be kept")
)

func IsWaitingForInputError(err error) bool {
	return errors.Is(err, ErrWaitingForInputResetInputs) || errors.Is(err, ErrWaitingForInputKeepInputs)
}
