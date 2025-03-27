package cycle

import (
	"errors"

	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
)

// Cycle contains the info about one activation cycle
type Cycle struct {
	*common.Chainable
	number            int
	activationResults *component.ActivationResultCollection
}

// New creates a new cycle
func New() *Cycle {
	return &Cycle{
		Chainable:         common.NewChainable(),
		activationResults: component.NewActivationResultCollection(),
	}
}

// ActivationResults getter
func (cycle *Cycle) ActivationResults() *component.ActivationResultCollection {
	return cycle.activationResults
}

// HasErrors tells whether the cycle is ended with activation errors (at lease one component returned an error)
func (cycle *Cycle) HasErrors() bool {
	return cycle.ActivationResults().HasErrors()
}

// AllErrorsCombined returns all errors occurred within the cycle as one error
func (cycle *Cycle) AllErrorsCombined() error {
	var allErrors error
	for _, ar := range cycle.ActivationResults().All() {
		if ar.IsError() {
			allErrors = errors.Join(allErrors, ar.ActivationError())
		}
	}

	return allErrors
}

// HasPanics tells whether the cycle is ended with panic (at lease one component panicked)
func (cycle *Cycle) HasPanics() bool {
	return cycle.ActivationResults().HasPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated
func (cycle *Cycle) HasActivatedComponents() bool {
	return cycle.ActivationResults().HasActivatedComponents()
}

// WithActivationResults adds multiple activation results
func (cycle *Cycle) WithActivationResults(activationResults ...*component.ActivationResult) *Cycle {
	cycle.activationResults = cycle.ActivationResults().Add(activationResults...)
	return cycle
}

// AddActivationResult adds a single activation result in a thread-safe way
func (cycle *Cycle) AddActivationResult(result *component.ActivationResult) *Cycle {
	cycle.activationResults = cycle.ActivationResults().Add(result)
	return cycle
}

// Number returns sequence number
func (cycle *Cycle) Number() int {
	return cycle.number
}

// WithNumber sets the sequence number
func (cycle *Cycle) WithNumber(number int) *Cycle {
	cycle.number = number
	return cycle
}

// WithErr returns cycle with error
func (cycle *Cycle) WithErr(err error) *Cycle {
	cycle.SetErr(err)
	return cycle
}
