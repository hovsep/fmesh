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
func (c *Cycle) ActivationResults() *component.ActivationResultCollection {
	return c.activationResults
}

// HasActivationErrors tells whether the cycle is ended with activation errors (at least one component returned an error).
func (c *Cycle) HasActivationErrors() bool {
	return c.ActivationResults().HasActivationErrors()
}

// AllErrorsCombined returns all errors occurred within the cycle as one error.
func (c *Cycle) AllErrorsCombined() error {
	var allErrors error
	for _, ar := range c.ActivationResults().AllAsMapOrNil() {
		if ar.IsError() {
			allErrors = errors.Join(allErrors, ar.ActivationErrorWithComponentName())
		}
	}

	return allErrors
}

// AllPanicsCombined returns all panics occurred within the cycle as one error.
func (c *Cycle) AllPanicsCombined() error {
	var allPanics error
	for _, ar := range c.ActivationResults().AllAsMapOrNil() {
		if ar.IsPanic() {
			allPanics = errors.Join(allPanics, ar.ActivationErrorWithComponentName())
		}
	}

	return allPanics
}

// HasActivationPanics tells whether the cycle is ended with panic (at lease one component panicked).
func (c *Cycle) HasActivationPanics() bool {
	return c.ActivationResults().HasActivationPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (c *Cycle) HasActivatedComponents() bool {
	return c.ActivationResults().HasActivatedComponents()
}

// WithActivationResults adds multiple activation results.
func (c *Cycle) WithActivationResults(activationResults ...*component.ActivationResult) *Cycle {
	c.activationResults = c.ActivationResults().With(activationResults...)
	return c
}

// AddActivationResult adds a single activation result in a thread-safe way.
func (c *Cycle) AddActivationResult(result *component.ActivationResult) *Cycle {
	c.activationResults = c.ActivationResults().With(result)
	return c
}

// Number returns sequence number.
func (c *Cycle) Number() int {
	return c.number
}

// WithNumber sets the sequence number.
func (c *Cycle) WithNumber(number int) *Cycle {
	c.number = number
	return c
}

// WithChainableErr sets a chainable error and returns the cycle.
func (c *Cycle) WithChainableErr(err error) *Cycle {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *Cycle) HasChainableErr() bool {
	return c.chainableErr != nil
}

// ChainableErr returns chainable error.
func (c *Cycle) ChainableErr() error {
	return c.chainableErr
}
