package cycle

import (
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
func (cycle *Cycle) ActivationResults() component.ActivationResultCollection {
	return cycle.activationResults
}

// HasErrors tells whether the cycle is ended wih activation errors (at lease one component returned an error)
func (cycle *Cycle) HasErrors() bool {
	return cycle.ActivationResults().HasErrors()
}

// HasPanics tells whether the cycle is ended wih panic(at lease one component panicked)
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

// Number returns sequence number
func (cycle *Cycle) Number() int {
	return cycle.number
}

// WithNumber sets the sequence number
func (cycle *Cycle) WithNumber(number int) *Cycle {
	cycle.number = number
	return cycle
}

// WithChainError returns cycle with error
func (cycle *Cycle) WithChainError(err error) *Cycle {
	cycle.SetChainError(err)
	return cycle
}
