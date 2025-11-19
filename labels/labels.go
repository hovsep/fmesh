package labels

import (
	"fmt"
)

// Collection provides safe access to labels with error handling.
type Collection struct {
	chainableErr error
	labels       Map
}

// NewCollection creates an initialized collection.
func NewCollection() *Collection {
	return &Collection{
		chainableErr: nil,
		labels:       make(Map),
	}
}

// All returns all labels as a map.
func (c *Collection) All() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.labels, nil
}

// AllMatch returns true if all labels in the collection satisfy the predicate.
func (c *Collection) AllMatch(pred LabelPredicate) bool {
	if c.HasChainableErr() {
		return false
	}

	for k, v := range c.labels {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any label in the collection satisfies the predicate.
func (c *Collection) AnyMatch(pred LabelPredicate) bool {
	if c.HasChainableErr() {
		return false
	}
	for k, v := range c.labels {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// CountMatch returns the number of labels that match the predicate.
func (c *Collection) CountMatch(pred LabelPredicate) int {
	if c.HasChainableErr() {
		return 0
	}
	count := 0
	for k, v := range c.labels {
		if pred(k, v) {
			count++
		}
	}
	return count
}

// Value returns the value of a single label or error if not found.
func (c *Collection) Value(label string) (string, error) {
	if c.HasChainableErr() {
		return "", c.ChainableErr()
	}

	value, ok := c.labels[label]
	if !ok {
		return "", fmt.Errorf("label %s not found", label)
	}

	return value, nil
}

// ValueOrDefault returns label value or default value in case of any error.
func (c *Collection) ValueOrDefault(label, defaultValue string) string {
	value, err := c.Value(label)
	if err != nil {
		return defaultValue
	}
	return value
}

// Add adds or updates a single label.
func (c *Collection) Add(label, value string) *Collection {
	if c.HasChainableErr() {
		return c
	}

	c.labels[label] = value
	return c
}

// AddMany adds or updates multiple labels.
func (c *Collection) AddMany(labels Map) *Collection {
	if c.HasChainableErr() {
		return c
	}
	for label, value := range labels {
		c.Add(label, value)
	}
	return c
}

// Without removes given labels.
func (c *Collection) Without(labels ...string) *Collection {
	if c.HasChainableErr() {
		return c
	}

	for _, label := range labels {
		delete(c.labels, label)
	}
	return c
}

// Has returns true when the collection has given label.
func (c *Collection) Has(label string) bool {
	if c.HasChainableErr() {
		return false
	}
	_, ok := c.labels[label]
	return ok
}

// HasAll checks if a collection has all given labels with disregard of their values.
func (c *Collection) HasAll(labels ...string) bool {
	if c.HasChainableErr() {
		return false
	}
	for _, label := range labels {
		if !c.Has(label) {
			return false
		}
	}
	return true
}

// HasAny checks if a collection has any of the given labels.
func (c *Collection) HasAny(labels ...string) bool {
	if c.HasChainableErr() {
		return false
	}
	for _, label := range labels {
		if c.Has(label) {
			return true
		}
	}
	return false
}

// ValueIs returns true when a collection has given label with a given value.
func (c *Collection) ValueIs(label, value string) bool {
	if c.HasChainableErr() {
		return false
	}

	if !c.Has(label) {
		return false
	}

	l, err := c.Value(label)
	if err != nil {
		return false
	}

	return l == value
}

// Len returns the number of labels.
func (c *Collection) Len() int {
	if c.HasChainableErr() {
		return 0
	}
	return len(c.labels)
}

// IsEmpty returns true when there are no labels in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}

// Clear removes all labels from the collection.
func (c *Collection) Clear() *Collection {
	if c.HasChainableErr() {
		return c
	}
	c.labels = make(Map)
	return c
}

// ForEach applies the action to each label and returns the collection for chaining.
func (c *Collection) ForEach(action func(label, value string)) *Collection {
	if c.HasChainableErr() {
		return c
	}
	for k, v := range c.labels {
		action(k, v)
	}
	return c
}

// Filter returns a new collection with labels that pass the predicate.
func (c *Collection) Filter(pred LabelPredicate) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	filtered := NewCollection()
	for k, v := range c.labels {
		if pred(k, v) {
			filtered.Add(k, v)
		}
	}
	return filtered
}

// Map transforms labels and returns a new collection.
func (c *Collection) Map(mapper LabelMapper) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	transformed := NewCollection()
	for k, v := range c.labels {
		newK, newV := mapper(k, v)
		transformed.Add(newK, newV)
	}
	return transformed
}

// WithChainableErr sets a chainable error and returns the collection.
func (c *Collection) WithChainableErr(err error) *Collection {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *Collection) HasChainableErr() bool {
	return c.chainableErr != nil
}

// HasAllFrom returns true if the current collection contains all labels
// present in the other collection (ignoring values). Returns false if either
// collection has a chainable error.
func (c *Collection) HasAllFrom(other *Collection) bool {
	if c.HasChainableErr() || other.HasChainableErr() {
		return false
	}

	if other.Len() > c.Len() {
		return false
	}

	return other.AllMatch(func(label, _ string) bool {
		return c.Has(label)
	})
}

// HasAnyFrom returns true if the current collection contains at least one
// label present in the other collection (ignoring values). Returns false if
// either collection has a chainable error.
func (c *Collection) HasAnyFrom(other *Collection) bool {
	if c.HasChainableErr() || other.HasChainableErr() {
		return false
	}

	if other.IsEmpty() || c.IsEmpty() {
		return false
	}

	return other.AnyMatch(func(label, _ string) bool {
		return c.Has(label)
	})
}

// ChainableErr returns the chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}
