package cycle

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"sync"
)

// Cycle contains the info about one activation cycle
type Cycle struct {
	sync.Mutex
	*common.Chainable
	number            int
	activationResults component.ActivationResultCollection
}

// New creates a new cycle
func New() *Cycle {
	return &Cycle{
		Chainable:         common.NewChainable(),
		activationResults: component.NewActivationResultCollection(),
	}
}

// ActivationResults getter
func (c *Cycle) ActivationResults() component.ActivationResultCollection {
	return c.activationResults
}

// HasErrors tells whether the cycle is ended wih activation errors (at lease one component returned an error)
func (c *Cycle) HasErrors() bool {
	return c.ActivationResults().HasErrors()
}

// ConsolidatedError returns all errors and panics occurred during activation cycle together as single error
func (c *Cycle) ConsolidatedError() error {
	var err error
	for componentName, activationResult := range c.ActivationResults() {
		if activationResult.IsError() || activationResult.IsPanic() {
			err = errors.Join(err, fmt.Errorf("component: %s : %w", componentName, activationResult.ActivationError()))
		}
	}

	return err
}

// HasPanics tells whether the cycle is ended wih panic(at lease one component panicked)
func (c *Cycle) HasPanics() bool {
	return c.ActivationResults().HasPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated
func (c *Cycle) HasActivatedComponents() bool {
	return c.ActivationResults().HasActivatedComponents()
}

// WithActivationResults adds multiple activation results
func (c *Cycle) WithActivationResults(activationResults ...*component.ActivationResult) *Cycle {
	c.activationResults = c.ActivationResults().Add(activationResults...)
	return c
}

// Number returns sequence number
func (c *Cycle) Number() int {
	return c.number
}

// WithNumber sets the sequence number
func (c *Cycle) WithNumber(number int) *Cycle {
	c.number = number
	return c
}

// WithErr returns cycle with chained error
func (c *Cycle) WithErr(err error) *Cycle {
	c.SetErr(err)
	return c
}
