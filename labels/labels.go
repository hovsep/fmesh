package labels

import (
	"fmt"
	"maps"
	"slices"
)

// Collection is a mutable key-value string store.
// All write methods modify the receiver in place.
type Collection struct {
	labels map[string]string
}

// NewCollection creates an initialized collection.
func NewCollection() *Collection {
	return &Collection{
		labels: make(map[string]string),
	}
}

// All returns all labels as a map (a defensive copy; mutating the returned map
// does not change the collection).
func (c *Collection) All() (map[string]string, error) {
	return maps.Clone(c.labels), nil
}

// Keys returns all label names as a sorted slice. The caller owns the returned slice.
func (c *Collection) Keys() []string {
	keys := make([]string, 0, len(c.labels))
	for k := range c.labels {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Values returns all label values as a slice sorted by their corresponding key. The caller owns the returned slice.
func (c *Collection) Values() []string {
	keys := c.Keys()
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = c.labels[k]
	}
	return values
}

// Every returns true if all labels in the collection satisfy the predicate.
// Returns true for an empty collection (vacuous truth).
func (c *Collection) Every(pred Predicate) bool {
	for k, v := range c.labels {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if any label in the collection satisfies the predicate.
func (c *Collection) Any(pred Predicate) bool {
	for k, v := range c.labels {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// Count returns the number of labels that match the predicate.
func (c *Collection) Count(pred Predicate) int {
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
	c.labels[label] = value
	return c
}

// AddMany adds or updates multiple labels.
func (c *Collection) AddMany(labels map[string]string) *Collection {
	for label, value := range labels {
		c.Add(label, value)
	}
	return c
}

// Merge returns a new collection containing all labels from both c and other.
// On key conflict, other's value wins. Neither c nor other is modified.
func (c *Collection) Merge(other *Collection) *Collection {
	merged := NewCollection()
	maps.Copy(merged.labels, c.labels)
	maps.Copy(merged.labels, other.labels)
	return merged
}

// Remove deletes given labels.
func (c *Collection) Remove(labels ...string) *Collection {
	for _, label := range labels {
		delete(c.labels, label)
	}
	return c
}

// Has returns true when the collection has given label.
func (c *Collection) Has(label string) bool {
	_, ok := c.labels[label]
	return ok
}

// HasAll checks if a collection has all given labels with disregard of their values.
func (c *Collection) HasAll(labels ...string) bool {
	for _, label := range labels {
		if !c.Has(label) {
			return false
		}
	}
	return true
}

// HasAny checks if a collection has any of the given labels.
func (c *Collection) HasAny(labels ...string) bool {
	return slices.ContainsFunc(labels, c.Has)
}

// ValueIs returns true when a collection has given label with a given value.
func (c *Collection) ValueIs(label, value string) bool {
	v, ok := c.labels[label]
	return ok && v == value
}

// Len returns the number of labels.
func (c *Collection) Len() int {
	return len(c.labels)
}

// IsEmpty returns true when there are no labels in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}

// Clear removes all labels from the collection.
func (c *Collection) Clear() *Collection {
	c.labels = make(map[string]string)
	return c
}

// ForEach applies the action to each label. Returns the first error encountered.
func (c *Collection) ForEach(action func(label, value string) error) error {
	for k, v := range c.labels {
		if err := action(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Filter returns a new collection with labels that pass the predicate.
func (c *Collection) Filter(pred Predicate) *Collection {
	filtered := NewCollection()
	for k, v := range c.labels {
		if pred(k, v) {
			filtered.Add(k, v)
		}
	}
	return filtered
}

// Map transforms labels and returns a new collection.
func (c *Collection) Map(mapper Mapper) *Collection {
	transformed := NewCollection()
	for k, v := range c.labels {
		newK, newV := mapper(k, v)
		transformed.Add(newK, newV)
	}
	return transformed
}

// HasAllFrom returns true if c contains all labels present in other (values ignored).
func (c *Collection) HasAllFrom(other *Collection) bool {
	if other.Len() > c.Len() {
		return false
	}
	return other.Every(func(label, _ string) bool {
		return c.Has(label)
	})
}

// HasAnyFrom returns true if c contains at least one label present in other (values ignored).
func (c *Collection) HasAnyFrom(other *Collection) bool {
	if other.IsEmpty() || c.IsEmpty() {
		return false
	}
	return other.Any(func(label, _ string) bool {
		return c.Has(label)
	})
}
