package cycle

import (
	"github.com/hovsep/fmesh/component"
	"sync"
)

// Result contains the information about activation cycle
type Result struct {
	sync.Mutex
	cycleNumber       uint
	activationResults component.ActivationResultCollection
}

// Results contains the results of several activation cycles
type Results []*Result

// NewResult creates a new cycle result
func NewResult() *Result {
	return &Result{
		activationResults: make(component.ActivationResultCollection),
	}
}

// NewResults creates a collection
func NewResults() Results {
	return make(Results, 0)
}

func (cycleResult *Result) SetCycleNumber(n uint) *Result {
	cycleResult.cycleNumber = n
	return cycleResult
}

// CycleNumber getter
func (cycleResult *Result) CycleNumber() uint {
	return cycleResult.cycleNumber
}

// ActivationResults getter
func (cycleResult *Result) ActivationResults() component.ActivationResultCollection {
	return cycleResult.activationResults
}

// HasErrors tells whether the cycle is ended wih activation errors (at lease one component returned an error)
func (cycleResult *Result) HasErrors() bool {
	return cycleResult.ActivationResults().HasErrors()
}

// HasPanics tells whether the cycle is ended wih panic(at lease one component panicked)
func (cycleResult *Result) HasPanics() bool {
	return cycleResult.ActivationResults().HasPanics()
}

// HasActivatedComponents tells when at least one component in the cycle has activated
func (cycleResult *Result) HasActivatedComponents() bool {
	return cycleResult.ActivationResults().HasActivatedComponents()
}

// Add adds cycle results to existing collection
func (cycleResults Results) Add(newCycleResults ...*Result) Results {
	for _, cycleResult := range newCycleResults {
		cycleResults = append(cycleResults, cycleResult)
	}
	return cycleResults
}

// WithActivationResults adds multiple activation results
func (cycleResult *Result) WithActivationResults(activationResults ...*component.ActivationResult) *Result {
	cycleResult.activationResults = cycleResult.ActivationResults().Add(activationResults...)
	return cycleResult
}
