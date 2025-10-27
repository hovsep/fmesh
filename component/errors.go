package component

import (
	"errors"
	"fmt"
)

var (
	errNotFound             = errors.New("component not found")
	errWaitingForInputs     = errors.New("component is waiting for some inputs")
	errWaitingForInputsKeep = fmt.Errorf("%w: do not clear input ports", errWaitingForInputs)
	// ErrNoComponentsInCollection is returned when a collection has no components.
	ErrNoComponentsInCollection = errors.New("no components in collection")
	// ErrNoComponentMatchesPredicate is returned when no component matches the predicate.
	ErrNoComponentMatchesPredicate     = errors.New("no component matches the predicate")
	errUnexpectedErrorGettingComponent = errors.New("unexpected error getting component")
)

// NewErrWaitForInputs returns respective error.
func NewErrWaitForInputs(keepInputs bool) error {
	if keepInputs {
		return errWaitingForInputsKeep
	}
	return errWaitingForInputs
}
