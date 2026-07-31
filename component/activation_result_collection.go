package component

import (
	"maps"
	"sync"
)

// ActivationResultCollection is a collection of activation results.
// Thread-safe for concurrent access during activation.
type ActivationResultCollection struct {
	mu                sync.RWMutex
	activationResults map[string]*ActivationResult
}

// NewActivationResultCollection creates an empty collection.
func NewActivationResultCollection() *ActivationResultCollection {
	return &ActivationResultCollection{
		activationResults: make(map[string]*ActivationResult),
	}
}

// Add adds multiple activation results and returns the collection.
func (c *ActivationResultCollection) Add(activationResults ...*ActivationResult) *ActivationResultCollection {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, activationResult := range activationResults {
		c.activationResults[activationResult.ComponentName()] = activationResult
	}
	return c
}

// HasActivationErrors tells whether the collection contains at least one activation result with error and respective code.
func (c *ActivationResultCollection) HasActivationErrors() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, ar := range c.activationResults {
		if ar.IsError() {
			return true
		}
	}
	return false
}

// HasActivationPanics tells whether the collection contains at least one activation result with panic and respective code.
func (c *ActivationResultCollection) HasActivationPanics() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, ar := range c.activationResults {
		if ar.IsPanic() {
			return true
		}
	}
	return false
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (c *ActivationResultCollection) HasActivatedComponents() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, ar := range c.activationResults {
		if ar.Activated() {
			return true
		}
	}
	return false
}

// ByName returns the activation result by component name.
func (c *ActivationResultCollection) ByName(name string) *ActivationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if result, ok := c.activationResults[name]; ok {
		return result
	}
	return nil
}

// All returns a shallow copy of all activation results as a map.
// A copy is returned so the caller cannot mutate the internal state.
func (c *ActivationResultCollection) All() map[string]*ActivationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*ActivationResult, len(c.activationResults))
	maps.Copy(result, c.activationResults)
	return result
}

// Len returns the number of activation results in the collection.
func (c *ActivationResultCollection) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.activationResults)
}

// Every returns true if all activation results match the predicate.
func (c *ActivationResultCollection) Every(predicate ResultPredicate) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, result := range c.activationResults {
		if !predicate(result) {
			return false
		}
	}
	return true
}

// ForEach applies the action to each activation result. Returns the first error encountered.
func (c *ActivationResultCollection) ForEach(action func(*ActivationResult) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, result := range c.activationResults {
		if err := action(result); err != nil {
			return err
		}
	}
	return nil
}
