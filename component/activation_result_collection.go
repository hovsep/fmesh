package component

import (
	"sync"
)

// ActivationResultCollection is a collection of activation results.
// Thread-safe for concurrent access during activation.
type ActivationResultCollection struct {
	mu                sync.RWMutex
	chainableErr      error
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
	if c.HasChainableErr() {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, activationResult := range activationResults {
		c.activationResults[activationResult.ComponentName()] = activationResult
	}
	return c
}

// Without removes activation results by component name and returns the collection.
func (c *ActivationResultCollection) Without(componentNames ...string) *ActivationResultCollection {
	if c.HasChainableErr() {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, name := range componentNames {
		delete(c.activationResults, name)
	}

	return c
}

// HasActivationErrors tells whether the collection contains at least one activation result with error and respective code.
func (c *ActivationResultCollection) HasActivationErrors() bool {
	if c.HasChainableErr() {
		return false
	}
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
	if c.HasChainableErr() {
		return false
	}
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
	if c.HasChainableErr() {
		return false
	}
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
	if c.HasChainableErr() {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if result, ok := c.activationResults[name]; ok {
		return result
	}
	return nil
}

// All returns all activation results as a map.
func (c *ActivationResultCollection) All() (map[string]*ActivationResult, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activationResults, nil
}

// Len returns the number of activation results in the collection.
func (c *ActivationResultCollection) Len() int {
	if c.HasChainableErr() {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.activationResults)
}

// IsEmpty returns true when there are no activation results in the collection.
func (c *ActivationResultCollection) IsEmpty() bool {
	return c.Len() == 0
}

// WithChainableErr sets a chainable error and returns the collection.
func (c *ActivationResultCollection) WithChainableErr(err error) *ActivationResultCollection {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *ActivationResultCollection) HasChainableErr() bool {
	return c.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (c *ActivationResultCollection) ChainableErr() error {
	return c.chainableErr
}

// AllMatch returns true if all activation results match the predicate.
func (c *ActivationResultCollection) AllMatch(predicate ActivationResultPredicate) bool {
	if c.HasChainableErr() {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, result := range c.activationResults {
		if !predicate(result) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any activation result matches the predicate.
func (c *ActivationResultCollection) AnyMatch(predicate ActivationResultPredicate) bool {
	if c.HasChainableErr() {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, result := range c.activationResults {
		if predicate(result) {
			return true
		}
	}
	return false
}

// CountMatch returns the number of activation results that match the predicate.
func (c *ActivationResultCollection) CountMatch(predicate ActivationResultPredicate) int {
	if c.HasChainableErr() {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, result := range c.activationResults {
		if predicate(result) {
			count++
		}
	}
	return count
}

// ForEach applies the action to each activation result and returns the collection for chaining.
func (c *ActivationResultCollection) ForEach(action func(*ActivationResult)) *ActivationResultCollection {
	if c.HasChainableErr() {
		return c
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, result := range c.activationResults {
		action(result)
	}
	return c
}

// Clear removes all activation results from the collection.
func (c *ActivationResultCollection) Clear() *ActivationResultCollection {
	if c.HasChainableErr() {
		return c
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activationResults = make(map[string]*ActivationResult)
	return c
}
