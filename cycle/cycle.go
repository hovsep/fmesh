package cycle

import (
	"errors"

	"github.com/hovsep/fmesh/component"
)

// Cycle contains the info about one activation cycle.
type Cycle struct {
	chainableErr      error
	number            int
	activationResults *component.ActivationResultCollection
}

// New creates a new cycle.
func New() *Cycle {
	return &Cycle{
		chainableErr:      nil,
		activationResults: component.NewActivationResultCollection(),
	}
}

// ActivationResults getter.
func (cycle *Cycle) ActivationResults() *component.ActivationResultCollection {
	return cycle.activationResults
}

// HasActivationErrors tells whether the cycle is ended with activation errors (at least one component returned an error).
func (cycle *Cycle) HasActivationErrors() bool {
	return cycle.ActivationResults().HasActivationErrors()
}

// AllErrorsCombined returns all errors occurred within the cycle as one error.
func (cycle *Cycle) AllErrorsCombined() error {
	var allErrors error
	for _, ar := range cycle.ActivationResults().All() {
		if ar.IsError() {
			allErrors = errors.Join(allErrors, ar.ActivationErrorWithComponentName())
		}
	}

	return allErrors
}

// AllPanicsCombined returns all panics occurred within the cycle as one error.
func (cycle *Cycle) AllPanicsCombined() error {
	var allPanics error
	for _, ar := range cycle.ActivationResults().All() {
		if ar.IsPanic() {
			allPanics = errors.Join(allPanics, ar.ActivationErrorWithComponentName())
		}
	}

	return allPanics
}

// HasActivationPanics tells whether the cycle is ended with panic (at lease one component panicked).
func (cycle *Cycle) HasActivationPanics() bool {
	return cycle.ActivationResults().HasActivationPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (cycle *Cycle) HasActivatedComponents() bool {
	return cycle.ActivationResults().HasActivatedComponents()
}

// WithActivationResults adds multiple activation results.
func (cycle *Cycle) WithActivationResults(activationResults ...*component.ActivationResult) *Cycle {
	cycle.activationResults = cycle.ActivationResults().Add(activationResults...)
	return cycle
}

// AddActivationResult adds a single activation result in a thread-safe way.
func (cycle *Cycle) AddActivationResult(result *component.ActivationResult) *Cycle {
	cycle.activationResults = cycle.ActivationResults().Add(result)
	return cycle
}

// Number returns sequence number.
func (cycle *Cycle) Number() int {
	return cycle.number
}

// WithNumber sets the sequence number.
func (cycle *Cycle) WithNumber(number int) *Cycle {
	cycle.number = number
	return cycle
}

// WithChainableErr sets a chainable error and returns the cycle.
func (cycle *Cycle) WithChainableErr(err error) *Cycle {
	cycle.chainableErr = err
	return cycle
}

// HasChainableErr returns true when a chainable error is set.
func (cycle *Cycle) HasChainableErr() bool {
	return cycle.chainableErr != nil
}

// ChainableErr returns chainable error.
func (cycle *Cycle) ChainableErr() error {
	return cycle.chainableErr
}
