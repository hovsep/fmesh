package cycle

import (
	"github.com/hovsep/fmesh/component"
	"sync"
)

// Result contains the information about activation cycle
type Result struct {
	sync.Mutex
	ActivationResults map[string]component.ActivationResult
}

// Results contains the results of several activation cycles
type Results []*Result

func NewResult() *Result {
	return &Result{
		ActivationResults: make(map[string]component.ActivationResult),
	}
}

func NewResults() Results {
	return make(Results, 0)
}

// HasErrors tells whether the cycle is ended wih activation errors
func (r *Result) HasErrors() bool {
	for _, ar := range r.ActivationResults {
		if ar.HasError() {
			return true
		}
	}
	return false
}

func (r *Result) HasActivatedComponents() bool {
	for _, ar := range r.ActivationResults {
		if ar.Activated() {
			return true
		}
	}
	return false
}
