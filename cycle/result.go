package cycle

import (
	"github.com/hovsep/fmesh/component"
	"sync"
)

// Result contains the information about activation cycle
type Result struct {
	sync.Mutex
	cycleNumber       uint
	activationResults component.ActivationResults
}

// Results contains the results of several activation cycles
type Results []*Result

// NewResult creates a new cycle result
func NewResult() *Result {
	return &Result{
		activationResults: make(component.ActivationResults),
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
func (cycleResult *Result) ActivationResults() component.ActivationResults {
	return cycleResult.activationResults
}

// WithActivationResults adds multiple activation results
func (cycleResult *Result) WithActivationResults(activationResults ...*component.ActivationResult) *Result {
	for _, activationResult := range activationResults {
		cycleResult.activationResults[activationResult.ComponentName()] = activationResult
	}
	return cycleResult
}

// HasErrors tells whether the cycle is ended wih activation errors (at lease one component returned an error)
func (cycleResult *Result) HasErrors() bool {
	for _, ar := range cycleResult.activationResults {
		if ar.HasError() {
			return true
		}
	}
	return false
}

// HasPanics tells whether the cycle is ended wih panic(at lease one component panicked)
func (cycleResult *Result) HasPanics() bool {
	for _, ar := range cycleResult.activationResults {
		if ar.HasPanic() {
			return true
		}
	}
	return false
}

func (cycleResult *Result) HasActivatedComponents() bool {
	for _, ar := range cycleResult.activationResults {
		if ar.Activated() {
			return true
		}
	}
	return false
}

// Add adds cycle results to existing collection
func (cycleResults Results) Add(newCycleResults ...*Result) Results {
	for _, cycleResult := range newCycleResults {
		cycleResults = append(cycleResults, cycleResult)
	}
	return cycleResults
}
