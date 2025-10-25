package labels

import (
	"fmt"
)

// Map is a map of labels.
type Map map[string]string

// Collection provides safe access to labels with error handling.
type Collection struct {
	chainableErr error
	labels       Map
}

// LabelPredicate tests a label key-value pair.
type LabelPredicate func(label, value string) bool

// NewCollection creates an initialized collection.
func NewCollection(labels Map) *Collection {
	if labels == nil {
		labels = make(Map)
	}
	return &Collection{
		chainableErr: nil,
		labels:       labels,
	}
}

// All returns true if all labels in the collection satisfy the predicate.
func (c *Collection) All(pred LabelPredicate) bool {
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

// With adds or updates a single label.
func (c *Collection) With(label, value string) *Collection {
	if c.HasChainableErr() {
		return c
	}

	c.labels[label] = value
	return c
}

// WithMany adds or updates multiple labels.
func (c *Collection) WithMany(labels Map) *Collection {
	if c.HasChainableErr() {
		return c
	}
	for label, value := range labels {
		c.With(label, value)
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

// MatchesAll returns true if the collection contains all key-value pairs from the given map.
func (c *Collection) MatchesAll(labels Map) bool {
	if c.HasChainableErr() {
		return false
	}
	for k, v := range labels {
		if !c.ValueIs(k, v) {
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

// MatchesAny returns true if the collection contains any key-value pair from the given map.
func (c *Collection) MatchesAny(labels Map) bool {
	if c.HasChainableErr() {
		return false
	}
	for k, v := range labels {
		if c.ValueIs(k, v) {
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

// WithChainableErr sets a chainable error and returns the collection.
func (c *Collection) WithChainableErr(err error) *Collection {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *Collection) HasChainableErr() bool {
	return c.chainableErr != nil
}

// ChainableErr returns chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}
