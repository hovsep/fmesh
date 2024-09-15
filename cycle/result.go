package cycle

import (
	"github.com/hovsep/fmesh/component"
	"sync"
)

// Result contains the information about activation cycle
type Result struct {
	sync.Mutex
	ActivationResults map[string]*component.ActivationResult
}

// Results contains the results of several activation cycles
type Results []*Result

func NewResult() *Result {
	return &Result{
		ActivationResults: make(map[string]*component.ActivationResult),
	}
}

// WithActivationResult adds an activation result of particular component to cycle result
func (result *Result) WithActivationResult(activationResult *component.ActivationResult) *Result {
	result.ActivationResults[activationResult.ComponentName()] = activationResult
	return result
}

// WithActivationResults adds multiple activation results
func (result *Result) WithActivationResults(activationResults ...*component.ActivationResult) *Result {
	for _, activationResult := range activationResults {
		result.WithActivationResult(activationResult)
	}
	return result
}

func NewResults() Results {
	return make(Results, 0)
}

// HasErrors tells whether the cycle is ended wih activation errors (at lease one component returned an error)
func (result *Result) HasErrors() bool {
	for _, ar := range result.ActivationResults {
		if ar.HasError() {
			return true
		}
	}
	return false
}

// HasPanics tells whether the cycle is ended wih panic(at lease one component panicked)
func (result *Result) HasPanics() bool {
	for _, ar := range result.ActivationResults {
		if ar.HasPanic() {
			return true
		}
	}
	return false
}

func (result *Result) HasActivatedComponents() bool {
	for _, ar := range result.ActivationResults {
		if ar.Activated() {
			return true
		}
	}
	return false
}
