package labels

import (
	"fmt"
	"maps"
	"slices"
)

// Collection is a mutable key-value string store.
// All write methods modify the receiver in place.
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

// All returns all labels as a map (a defensive copy; mutating the returned map
// does not change the collection).
func (c *Collection) All() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return maps.Clone(c.labels), nil
}

// Keys returns all label names as a sorted slice. The caller owns the returned slice.
func (c *Collection) Keys() []string {
	if c.HasChainableErr() {
		return nil
	}
	keys := make([]string, 0, len(c.labels))
	for k := range c.labels {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Values returns all label values as a slice sorted by their corresponding key. The caller owns the returned slice.
func (c *Collection) Values() []string {
	if c.HasChainableErr() {
		return nil
	}
	keys := c.Keys()
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = c.labels[k]
	}
	return values
}

// Every returns true if all labels in the collection satisfy the predicate.
// Returns true for an empty collection (vacuous truth).
func (c *Collection) Every(pred LabelPredicate) bool {
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

// Any returns true if any label in the collection satisfies the predicate.
func (c *Collection) Any(pred LabelPredicate) bool {
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

// Count returns the number of labels that match the predicate.
func (c *Collection) Count(pred LabelPredicate) int {
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

// Merge returns a new collection containing all labels from both c and other.
// On key conflict, other's value wins. Neither c nor other is modified.
func (c *Collection) Merge(other *Collection) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	if other.HasChainableErr() {
		return NewCollection().WithChainableErr(other.ChainableErr())
	}
	merged := NewCollection()
	maps.Copy(merged.labels, c.labels)
	maps.Copy(merged.labels, other.labels)
	return merged
}

// Remove deletes given labels.
func (c *Collection) Remove(labels ...string) *Collection {
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
	return slices.ContainsFunc(labels, c.Has)
}

// ValueIs returns true when a collection has given label with a given value.
func (c *Collection) ValueIs(label, value string) bool {
	if c.HasChainableErr() {
		return false
	}
	v, ok := c.labels[label]
	return ok && v == value
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
func (c *Collection) ForEach(action func(label, value string) error) *Collection {
	if c.HasChainableErr() {
		return c
	}
	for k, v := range c.labels {
		if err := action(k, v); err != nil {
			c.chainableErr = err
			return c
		}
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

// HasAllFrom returns true if c contains all labels present in other (values ignored).
func (c *Collection) HasAllFrom(other *Collection) bool {
	if c.HasChainableErr() || other.HasChainableErr() {
		return false
	}

	if other.Len() > c.Len() {
		return false
	}

	return other.Every(func(label, _ string) bool {
		return c.Has(label)
	})
}

// HasAnyFrom returns true if c contains at least one label present in other (values ignored).
func (c *Collection) HasAnyFrom(other *Collection) bool {
	if c.HasChainableErr() || other.HasChainableErr() {
		return false
	}

	if other.IsEmpty() || c.IsEmpty() {
		return false
	}

	return other.Any(func(label, _ string) bool {
		return c.Has(label)
	})
}

// ChainableErr returns the chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}
