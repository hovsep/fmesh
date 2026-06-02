package cycle

import (
	"errors"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/meta"
)

// Cycle contains the info about one activation cycle.
type Cycle struct {
	number            int
	labels            *meta.Labels
	scalars           *meta.Scalars
	activationResults *component.ActivationResultCollection
}

// New creates a new cycle.
func New() *Cycle {
	return &Cycle{
		labels:            meta.NewLabels(),
		scalars:           meta.NewScalars(),
		activationResults: component.NewActivationResultCollection(),
	}
}

// ActivationResults returns the cycle's activation results collection.
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
	activationResults := c.ActivationResults().All()
	for _, ar := range activationResults {
		if ar.IsError() {
			allErrors = errors.Join(allErrors, ar.ActivationErrorWithComponentName())
		}
	}

	return allErrors
}

// AllPanicsCombined returns all panics occurred within the cycle as one error.
func (c *Cycle) AllPanicsCombined() error {
	var allPanics error
	activationResults := c.ActivationResults().All()
	for _, ar := range activationResults {
		if ar.IsPanic() {
			allPanics = errors.Join(allPanics, ar.ActivationErrorWithComponentName())
		}
	}

	return allPanics
}

// HasActivationPanics tells whether the cycle ended with at least one component panicking.
func (c *Cycle) HasActivationPanics() bool {
	return c.ActivationResults().HasActivationPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (c *Cycle) HasActivatedComponents() bool {
	return c.ActivationResults().HasActivatedComponents()
}

// AddActivationResults adds multiple activation results.
func (c *Cycle) AddActivationResults(activationResults ...*component.ActivationResult) *Cycle {
	c.activationResults = c.ActivationResults().Add(activationResults...)
	return c
}

// Number returns sequence number.
func (c *Cycle) Number() int {
	return c.number
}

// SetNumber sets the sequence number.
func (c *Cycle) SetNumber(number int) *Cycle {
	c.number = number
	return c
}

// Labels returns the cycle's labels store.
func (c *Cycle) Labels() *meta.Labels {
	return c.labels
}

// AddLabel adds or updates a single label.
func (c *Cycle) AddLabel(name, value string) *Cycle {
	c.labels.Set(name, value)
	return c
}

// Scalars returns the cycle's scalars store.
func (c *Cycle) Scalars() *meta.Scalars {
	return c.scalars
}

// AddScalar adds or updates a single scalar.
func (c *Cycle) AddScalar(name string, value float64) *Cycle {
	c.scalars.Set(name, value)
	return c
}
