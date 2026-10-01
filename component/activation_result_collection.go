package component

import (
	"maps"
	"slices"
	"sync"
)

// ActivationResultCollection is a collection of activation results.
// Thread-safe for concurrent access during activation. Traversals walk a component-name-ordered
// snapshot, so callbacks run without the lock held.
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

// Remove removes activation results by component name and returns the collection.
func (c *ActivationResultCollection) Remove(componentNames ...string) *ActivationResultCollection {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, name := range componentNames {
		delete(c.activationResults, name)
	}

	return c
}

// HasActivationErrors tells whether the collection contains at least one activation result with error and respective code.
func (c *ActivationResultCollection) HasActivationErrors() bool {
	return c.Any((*ActivationResult).IsError)
}

// HasActivationPanics tells whether the collection contains at least one activation result with panic and respective code.
func (c *ActivationResultCollection) HasActivationPanics() bool {
	return c.Any((*ActivationResult).IsPanic)
}

// HasActivatedComponents tells when at least one component in the cycle has activated.
func (c *ActivationResultCollection) HasActivatedComponents() bool {
	return c.Any((*ActivationResult).Activated)
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

// AllOrdered returns the results sorted by component name — the order to use
// for anything rendered, so that output does not follow map order.
func (c *ActivationResultCollection) AllOrdered() []*ActivationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ordered := make([]*ActivationResult, 0, len(c.activationResults))
	for _, name := range slices.Sorted(maps.Keys(c.activationResults)) {
		ordered = append(ordered, c.activationResults[name])
	}
	return ordered
}

// Len returns the number of activation results in the collection.
func (c *ActivationResultCollection) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.activationResults)
}

// IsEmpty returns true when there are no activation results in the collection.
func (c *ActivationResultCollection) IsEmpty() bool {
	return c.Len() == 0
}

// Every returns true if all activation results match the predicate.
func (c *ActivationResultCollection) Every(predicate ResultPredicate) bool {
	for _, result := range c.AllOrdered() {
		if !predicate(result) {
			return false
		}
	}
	return true
}

// Any returns true if any activation result matches the predicate.
func (c *ActivationResultCollection) Any(predicate ResultPredicate) bool {
	return slices.ContainsFunc(c.AllOrdered(), predicate)
}

// Count returns the number of activation results that match the predicate.
func (c *ActivationResultCollection) Count(predicate ResultPredicate) int {
	count := 0
	for _, result := range c.AllOrdered() {
		if predicate(result) {
			count++
		}
	}
	return count
}

// ForEach applies the action to each activation result in component-name order and returns the first error.
// It walks a snapshot, so the action may change the collection.
func (c *ActivationResultCollection) ForEach(action func(*ActivationResult) error) error {
	for _, result := range c.AllOrdered() {
		if err := action(result); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all activation results from the collection.
func (c *ActivationResultCollection) Clear() *ActivationResultCollection {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activationResults = make(map[string]*ActivationResult)
	return c
}

// FindAny returns the first activation result matching the predicate, in component-name order, or nil.
func (c *ActivationResultCollection) FindAny(predicate ResultPredicate) *ActivationResult {
	for _, ar := range c.AllOrdered() {
		if predicate(ar) {
			return ar
		}
	}
	return nil
}

// Filter returns a new collection with activation results that match the predicate.
func (c *ActivationResultCollection) Filter(predicate ResultPredicate) *ActivationResultCollection {
	filtered := NewActivationResultCollection()
	for _, ar := range c.AllOrdered() {
		if predicate(ar) {
			filtered.Add(ar)
		}
	}
	return filtered
}
