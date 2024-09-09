package hop

import "sync"

// HopResult describes the outcome of every single component activation in single hop
type HopResult struct {
	sync.Mutex
	ActivationResults map[string]error
}

func (r *HopResult) HasErrors() bool {
	for _, err := range r.ActivationResults {
		if err != nil {
			return true
		}
	}
	return false
}
