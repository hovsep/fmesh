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
	return c.joinActivationErrors((*component.ActivationResult).IsError)
}

// AllPanicsCombined returns all panics occurred within the cycle as one error.
func (c *Cycle) AllPanicsCombined() error {
	return c.joinActivationErrors((*component.ActivationResult).IsPanic)
}

// joinActivationErrors joins the errors of every activation result the predicate
// matches, each tagged with the component that produced it. Returns nil when
// nothing matches.
func (c *Cycle) joinActivationErrors(matching component.ResultPredicate) error {
	var joined error
	for _, activationResult := range c.ActivationResults().All() {
		if matching(activationResult) {
			joined = errors.Join(joined, activationResult.ActivationErrorWithComponentName())
		}
	}
	return joined
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
// cycle did so only to say it is waiting for inputs — half the livelock test, the
// other half being that no signal moved (see FMesh.detectLivelock).
//
// Components that did not activate are not recorded at all, so an idle mesh does
// not read as stalled.
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
