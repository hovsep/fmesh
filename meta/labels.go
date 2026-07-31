package meta

import (
	"fmt"
	"slices"
)

// Labels is a mutable key-value string store.
// All write methods modify the receiver in place.
type Labels struct {
	store[string]
}

// NewLabels creates an initialized Labels store.
func NewLabels() *Labels {
	c := &Labels{}
	c.init()
	return c
}

// Set adds or updates a single label (upsert semantics).
func (c *Labels) Set(label, value string) *Labels {
	c.set(label, value)
	return c
}

// SetMany adds or updates multiple labels (upsert semantics).
func (c *Labels) SetMany(labels map[string]string) *Labels {
	c.setMany(labels)
	return c
}

// Remove deletes the named labels. Missing names are silently ignored.
func (c *Labels) Remove(labels ...string) *Labels {
	c.remove(labels...)
	return c
}

// Clear removes every label.
func (c *Labels) Clear() *Labels {
	c.clear()
	return c
}

// Value returns the value of a single label, or an error if not found.
func (c *Labels) Value(label string) (string, error) {
	v, ok := c.lookup(label)
	if !ok {
		return "", fmt.Errorf("label %s not found", label)
	}
	return v, nil
}

// Values returns all label values as a slice sorted by their corresponding key.
// The caller owns the returned slice.
func (c *Labels) Values() []string {
	keys := c.Keys()
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = c.entries[k]
	}
	return values
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

// Merge returns a new collection containing all labels from both c and other.
// On key conflict, other's value wins. Neither c nor other is modified.
func (c *Labels) Merge(other *Labels) *Labels {
	merged := NewLabels()
	c.mergeInto(merged.entries, other.entries)
	return merged
}

// Filter returns a new collection with labels that pass the predicate.
func (c *Labels) Filter(pred Predicate) *Labels {
	filtered := NewLabels()
	c.filterInto(filtered.entries, pred)
	return filtered
}

// Map transforms labels and returns a new collection.
func (c *Labels) Map(mapper Mapper) *Labels {
	transformed := NewLabels()
	for k, v := range c.entries {
		newK, newV := mapper(k, v)
		transformed.entries[newK] = newV
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
