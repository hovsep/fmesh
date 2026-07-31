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

// Labels returns the cycle's labels store.
func (c *Cycle) Labels() *meta.Labels {
	return c.labels
}

// Scalars returns the cycle's scalars store.
func (c *Cycle) Scalars() *meta.Scalars {
	return c.scalars
}

// HasActivationErrors tells whether the cycle is ended with activation errors (at least one component returned an error).
func (c *Cycle) HasActivationErrors() bool {
	return c.ActivationResults().HasActivationErrors()
}

// AllErrorsCombined returns all errors occurred within the cycle as one error.
func (c *Cycle) AllErrorsCombined() error {
	return c.combine((*component.ActivationResult).IsError)
}

// AllPanicsCombined returns all panics occurred within the cycle as one error.
func (c *Cycle) AllPanicsCombined() error {
	return c.combine((*component.ActivationResult).IsPanic)
}

// combine joins the activation errors of every result the predicate selects.
func (c *Cycle) combine(selects func(*component.ActivationResult) bool) error {
	var combined error
	for _, ar := range c.ActivationResults().All() {
		if selects(ar) {
			combined = errors.Join(combined, ar.ActivationErrorWithComponentName())
		}
	}
	return combined
}

// HasActivationPanics tells whether the cycle ended with at least one component panicking.
func (c *Cycle) HasActivationPanics() bool {
	return c.ActivationResults().HasActivationPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (c *Cycle) HasActivatedComponents() bool {
	return c.ActivationResults().HasActivatedComponents()
}

// AllActivatedAreWaiting reports whether every component that activated in this
// cycle did so only to say it is waiting for inputs.
//
// This is the shape of a stalled cycle. Waiting components are never drained, so
// such a cycle moves no signals at all: whatever the mesh was holding when the
// cycle began, it is still holding. On its own that does not prove a livelock —
// a component accumulating input is legitimately waiting — which is why the run
// loop pairs this with a check that the signals really did not move.
//
// Components that did not activate at all are not recorded (see runCycle), so an
// idle component does not make a cycle look stalled. A result that is anything
// other than waiting — an error, a panic, a hook failure, a plain success —
// makes this false.
func (c *Cycle) AllActivatedAreWaiting() bool {
	if !c.HasActivatedComponents() {
		return false
	}
	return c.ActivationResults().Every(component.IsWaitingForInput)
}

// AddActivationResults adds multiple activation results.
// Safe for concurrent use: the underlying collection is mutex-protected and
// the field is never reassigned.
func (c *Cycle) AddActivationResults(activationResults ...*component.ActivationResult) *Cycle {
	c.ActivationResults().Add(activationResults...)
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
