package component

import (
	"cmp"
	"slices"
	"sync"
)

// ActivationResultCollection is a collection of activation results.
// Thread-safe for concurrent access during activation. ForEach and Find walk a
// component-name-ordered snapshot, so their callbacks run without the lock held.
// The order-independent queries (Any, Every, Count, Filter and the Has* methods)
// read under the lock and must not change the collection from their predicate.
type ActivationResultCollection struct {
	mu sync.RWMutex
	// results is sorted by component name, one entry per component.
	results []*ActivationResult
}

// NewActivationResultCollection creates an empty collection.
func NewActivationResultCollection() *ActivationResultCollection {
	return &ActivationResultCollection{}
}

func compareName(e *ActivationResult, name string) int { return cmp.Compare(e.componentName, name) }

func byName(a, b *ActivationResult) int { return cmp.Compare(a.componentName, b.componentName) }

// Add adds multiple activation results and returns the collection. A result for
// a component already in the collection, or repeated later in the same call,
// replaces the earlier one.
func (c *ActivationResultCollection) Add(activationResults ...*ActivationResult) *ActivationResultCollection {
	c.mu.Lock()
	defer c.mu.Unlock()

	// The run loop adds a whole cycle at once into an empty collection, already
	// in name order: that takes one copy, not an insert per result.
	if len(c.results) == 0 {
		c.results = sortedByName(activationResults)
		return c
	}

	for _, ar := range activationResults {
		i, found := slices.BinarySearchFunc(c.results, ar.componentName, compareName)
		if found {
			c.results[i] = ar
			continue
		}
		c.results = slices.Insert(c.results, i, ar)
	}
	return c
}

// sortedByName returns a name-sorted copy of results, keeping the last result
// of each name.
func sortedByName(results []*ActivationResult) []*ActivationResult {
	sorted := slices.Clone(results)
	if slices.IsSortedFunc(sorted, byName) && !hasAdjacentDuplicates(sorted) {
		return sorted
	}
	slices.SortStableFunc(sorted, byName)
	unique := sorted[:0]
	for i, ar := range sorted {
		if i+1 < len(sorted) && sorted[i+1].componentName == ar.componentName {
			continue
		}
		unique = append(unique, ar)
	}
	clear(sorted[len(unique):])
	return unique
}

func hasAdjacentDuplicates(sorted []*ActivationResult) bool {
	for i := 1; i < len(sorted); i++ {
		if sorted[i].componentName == sorted[i-1].componentName {
			return true
		}
	}
	return false
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
	i, found := slices.BinarySearchFunc(c.results, name, compareName)
	if !found {
		return nil
	}
	return c.results[i]
}

// All returns a shallow copy of all activation results as a map.
// A copy is returned so the caller cannot mutate the internal state.
func (c *ActivationResultCollection) All() map[string]*ActivationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*ActivationResult, len(c.results))
	for _, ar := range c.results {
		result[ar.componentName] = ar
	}
	return result
}

// AllOrdered returns the results sorted by component name — the order to use
// for anything rendered, so that output does not follow map order.
func (c *ActivationResultCollection) AllOrdered() []*ActivationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Clone(c.results)
}

// Len returns the number of activation results in the collection.
func (c *ActivationResultCollection) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.results)
}

// IsEmpty returns true when there are no activation results in the collection.
func (c *ActivationResultCollection) IsEmpty() bool {
	return c.Len() == 0
}

// Every returns true if all activation results match the predicate.
func (c *ActivationResultCollection) Every(predicate ResultPredicate) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, result := range c.results {
		if !predicate(result) {
			return false
		}
	}
	return true
}

// Any returns true if any activation result matches the predicate.
func (c *ActivationResultCollection) Any(predicate ResultPredicate) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.ContainsFunc(c.results, predicate)
}

// Count returns the number of activation results that match the predicate.
func (c *ActivationResultCollection) Count(predicate ResultPredicate) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := 0
	for _, result := range c.results {
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

// Find returns the first activation result matching the predicate, in component-name order, or nil.
func (c *ActivationResultCollection) Find(predicate ResultPredicate) *ActivationResult {
	for _, ar := range c.AllOrdered() {
		if predicate(ar) {
			return ar
		}
	}
	return nil
}

// Filter returns a new collection with activation results that match the predicate.
func (c *ActivationResultCollection) Filter(predicate ResultPredicate) *ActivationResultCollection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	filtered := NewActivationResultCollection()
	for _, ar := range c.results {
		if predicate(ar) {
			filtered.results = append(filtered.results, ar)
		}
	}
	return filtered
}
