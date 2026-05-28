package meta

import (
	"fmt"
	"maps"
	"slices"
)

// Labels is a mutable key-value string store.
// All write methods modify the receiver in place.
type Labels struct {
	labels map[string]string
}

// NewLabels creates an initialized Labels store.
func NewLabels() *Labels {
	return &Labels{
		labels: make(map[string]string),
	}
}

// All returns all labels as a map (a defensive copy; mutating the returned map
// does not change the collection).
func (c *Labels) All() map[string]string {
	return maps.Clone(c.labels)
}

// Keys returns all label names as a sorted slice. The caller owns the returned slice.
func (c *Labels) Keys() []string {
	keys := make([]string, 0, len(c.labels))
	for k := range c.labels {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Values returns all label values as a slice sorted by their corresponding key. The caller owns the returned slice.
func (c *Labels) Values() []string {
	keys := c.Keys()
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = c.labels[k]
	}
	return values
}

// Every returns true if all labels in the collection satisfy the predicate.
// Returns true for an empty collection (vacuous truth).
func (c *Labels) Every(pred Predicate) bool {
	for k, v := range c.labels {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if any label in the collection satisfies the predicate.
func (c *Labels) Any(pred Predicate) bool {
	for k, v := range c.labels {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// Count returns the number of labels that match the predicate.
func (c *Labels) Count(pred Predicate) int {
	count := 0
	for k, v := range c.labels {
		if pred(k, v) {
			count++
		}
	}
	return count
}

// Value returns the value of a single label or error if not found.
func (c *Labels) Value(label string) (string, error) {
	value, ok := c.labels[label]
	if !ok {
		return "", fmt.Errorf("label %s not found", label)
	}
	return value, nil
}

// ValueOrDefault returns label value or default value in case of any error.
func (c *Labels) ValueOrDefault(label, defaultValue string) string {
	value, err := c.Value(label)
	if err != nil {
		return defaultValue
	}
	return value
}

// Set adds or updates a single label (upsert semantics).
func (c *Labels) Set(label, value string) *Labels {
	c.labels[label] = value
	return c
}

// SetMany adds or updates multiple labels (upsert semantics).
func (c *Labels) SetMany(labels map[string]string) *Labels {
	for label, value := range labels {
		c.Set(label, value)
	}
	return c
}

// Merge returns a new collection containing all labels from both c and other.
// On key conflict, other's value wins. Neither c nor other is modified.
func (c *Labels) Merge(other *Labels) *Labels {
	merged := NewLabels()
	maps.Copy(merged.labels, c.labels)
	maps.Copy(merged.labels, other.labels)
	return merged
}

// Remove deletes given labels.
func (c *Labels) Remove(labels ...string) *Labels {
	for _, label := range labels {
		delete(c.labels, label)
	}
	return c
}

// Has returns true when the collection has given label.
func (c *Labels) Has(label string) bool {
	_, ok := c.labels[label]
	return ok
}

// HasAll checks if a collection has all given labels with disregard of their values.
func (c *Labels) HasAll(labels ...string) bool {
	for _, label := range labels {
		if !c.Has(label) {
			return false
		}
	}
	return true
}

// HasAny checks if a collection has any of the given labels.
func (c *Labels) HasAny(labels ...string) bool {
	return slices.ContainsFunc(labels, c.Has)
}

// ValueIs returns true when a collection has given label with a given value.
func (c *Labels) ValueIs(label, value string) bool {
	v, ok := c.labels[label]
	return ok && v == value
}

// Len returns the number of labels.
func (c *Labels) Len() int {
	return len(c.labels)
}

// IsEmpty returns true when there are no labels in the collection.
func (c *Labels) IsEmpty() bool {
	return c.Len() == 0
}

// Clear removes all labels from the collection.
func (c *Labels) Clear() *Labels {
	c.labels = make(map[string]string)
	return c
}

// ForEach applies the action to each label. Returns the first error encountered.
func (c *Labels) ForEach(action func(label, value string) error) error {
	for k, v := range c.labels {
		if err := action(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Filter returns a new collection with labels that pass the predicate.
func (c *Labels) Filter(pred Predicate) *Labels {
	filtered := NewLabels()
	for k, v := range c.labels {
		if pred(k, v) {
			filtered.Set(k, v)
		}
	}
	return filtered
}

// Map transforms labels and returns a new collection.
func (c *Labels) Map(mapper Mapper) *Labels {
	transformed := NewLabels()
	for k, v := range c.labels {
		newK, newV := mapper(k, v)
		transformed.Set(newK, newV)
	}
	return transformed
}

// HasAllFrom returns true if c contains all labels present in other (values ignored).
func (c *Labels) HasAllFrom(other *Labels) bool {
	if other.Len() > c.Len() {
		return false
	}
	return other.Every(func(label, _ string) bool {
		return c.Has(label)
	})
}

// HasAnyFrom returns true if c contains at least one label present in other (values ignored).
func (c *Labels) HasAnyFrom(other *Labels) bool {
	if other.IsEmpty() || c.IsEmpty() {
		return false
	}
	return other.Any(func(label, _ string) bool {
		return c.Has(label)
	})
}
