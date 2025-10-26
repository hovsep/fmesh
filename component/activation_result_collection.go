package component

import "sync"

// ActivationResultCollection is a collection.
type ActivationResultCollection struct {
	mu                sync.Mutex
	activationResults map[string]*ActivationResult
}

// NewActivationResultCollection creates empty collection.
func NewActivationResultCollection() *ActivationResultCollection {
	return &ActivationResultCollection{
		activationResults: make(map[string]*ActivationResult),
	}
}

// Add adds multiple activation results.
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
	for _, ar := range c.activationResults {
		if ar.IsError() {
			return true
		}
	}
	return false
}

// HasActivationPanics tells whether the collection contains at least one activation result with panic and respective code.
func (c *ActivationResultCollection) HasActivationPanics() bool {
	for _, ar := range c.activationResults {
		if ar.IsPanic() {
			return true
		}
	}
	return false
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (c *ActivationResultCollection) HasActivatedComponents() bool {
	for _, ar := range c.activationResults {
		if ar.Activated() {
			return true
		}
	}
	return false
}

// ByComponentName returns the activation result of given component.
func (c *ActivationResultCollection) ByComponentName(componentName string) *ActivationResult {
	if result, ok := c.activationResults[componentName]; ok {
		return result
	}

	return nil
}

// All returns all activation results.
func (c *ActivationResultCollection) All() map[string]*ActivationResult {
	return c.activationResults
}

// Len returns the number of activation results in the collection.
func (c *ActivationResultCollection) Len() int {
	return len(c.activationResults)
}
