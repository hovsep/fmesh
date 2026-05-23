package signal

import (
	"maps"
	"slices"

	"github.com/hovsep/fmesh/labels"
)

// Signal is a wrapper around the data flowing between components.
// Mutating-style methods return a new *Signal; receivers are never modified.
type Signal struct {
	labels  *labels.Collection
	payload []any // Slice is used in order to support nil payload
}

// cloneLabels returns an independent labels.Collection with the same entries.
func cloneLabels(c *labels.Collection) *labels.Collection {
	if c == nil {
		return labels.NewCollection()
	}
	return c.Filter(func(string, string) bool { return true })
}

// cloneSignal returns a copy of s with an independent labels collection.
// Payload is shallow-copied.
func cloneSignal(s *Signal) *Signal {
	if s == nil {
		return nil
	}
	return &Signal{
		labels:  cloneLabels(s.labels),
		payload: slices.Clone(s.payload),
	}
}

// New creates a new signal with the given payload.
func New(payload any) *Signal {
	return &Signal{
		labels:  labels.NewCollection(),
		payload: []any{payload},
	}
}

// Labels returns a defensive copy of the signal's labels collection.
func (s *Signal) Labels() *labels.Collection {
	return cloneLabels(s.labels)
}

// WithOnlyLabels replaces all labels and returns a new signal.
func (s *Signal) WithOnlyLabels(labelMap labels.Map) *Signal {
	next := cloneSignal(s)
	next.labels = labels.NewCollection().AddMany(maps.Clone(labelMap))
	return next
}

// WithLabels adds or updates labels and returns a new signal.
func (s *Signal) WithLabels(labelMap labels.Map) *Signal {
	next := cloneSignal(s)
	next.labels = next.labels.AddMany(maps.Clone(labelMap))
	return next
}

// WithLabel adds or updates a single label and returns a new signal.
func (s *Signal) WithLabel(name, value string) *Signal {
	next := cloneSignal(s)
	next.labels = next.labels.Add(name, value)
	return next
}

// WithNoLabels removes all labels and returns a new signal.
func (s *Signal) WithNoLabels() *Signal {
	next := cloneSignal(s)
	next.labels = labels.NewCollection()
	return next
}

// WithoutLabels removes specific labels and returns a new signal.
func (s *Signal) WithoutLabels(names ...string) *Signal {
	next := cloneSignal(s)
	next.labels = next.labels.Remove(names...)
	return next
}

// MapPayload applies a mapper function to the signal's payload and returns a new signal.
// The new signal preserves all labels from the original signal.
func (s *Signal) MapPayload(mapper PayloadMapper) *Signal {
	payload, _ := s.Payload()
	out := New(mapper(payload))
	out.labels = cloneLabels(s.labels)
	return out
}

// Payload returns the signal's payload. The value is shallow: if the payload is
// a pointer, slice, or map, the caller must not mutate it.
func (s *Signal) Payload() (any, error) {
	return s.payload[0], nil
}

// PayloadOrNil returns payload or nil in case of error.
func (s *Signal) PayloadOrNil() any {
	return s.PayloadOrDefault(nil)
}

// PayloadOrDefault returns payload or provided default value in case of error.
func (s *Signal) PayloadOrDefault(defaultPayload any) any {
	payload, err := s.Payload()
	if err != nil {
		return defaultPayload
	}
	return payload
}

// Map applies a given mapper func and returns a new signal.
func (s *Signal) Map(m Mapper) *Signal {
	return m(cloneSignal(s))
}
