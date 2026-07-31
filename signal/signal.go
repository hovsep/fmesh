package signal

import (
	"maps"
	"slices"

	"github.com/hovsep/fmesh/meta"
)

// Signal is a wrapper around the data flowing between components.
// Mutating-style methods return a new *Signal; receivers are never modified.
type Signal struct {
	labels  *meta.Labels
	scalars *meta.Scalars
	payload []any // Slice is used in order to support nil payload
}

// cloneLabels returns an independent meta.Labels with the same entries.
func cloneLabels(c *meta.Labels) *meta.Labels {
	if c == nil {
		return meta.NewLabels()
	}
	return c.Filter(func(string, string) bool { return true })
}

// cloneScalars returns an independent meta.Scalars with the same entries.
func cloneScalars(s *meta.Scalars) *meta.Scalars {
	if s == nil {
		return meta.NewScalars()
	}
	return s.Filter(func(string, float64) bool { return true })
}

// cloneSignal returns a copy of s with an independent labels and scalars collection.
// Payload is shallow-copied.
func cloneSignal(s *Signal) *Signal {
	if s == nil {
		return nil
	}
	return &Signal{
		labels:  cloneLabels(s.labels),
		scalars: cloneScalars(s.scalars),
		payload: slices.Clone(s.payload),
	}
}

// New creates a new signal with the given payload.
func New(payload any) *Signal {
	return &Signal{
		labels:  meta.NewLabels(),
		scalars: meta.NewScalars(),
		payload: []any{payload},
	}
}

// Labels returns a defensive copy of the signal's labels collection.
func (s *Signal) Labels() *meta.Labels {
	return cloneLabels(s.labels)
}

// WithOnlyLabels replaces all labels and returns a new signal.
func (s *Signal) WithOnlyLabels(labelMap map[string]string) *Signal {
	next := cloneSignal(s)
	next.labels = meta.NewLabels().SetMany(maps.Clone(labelMap))
	return next
}

// WithLabels adds or updates labels and returns a new signal.
func (s *Signal) WithLabels(labelMap map[string]string) *Signal {
	next := cloneSignal(s)
	next.labels = next.labels.SetMany(maps.Clone(labelMap))
	return next
}

// WithLabel adds or updates a single label and returns a new signal.
func (s *Signal) WithLabel(name, value string) *Signal {
	next := cloneSignal(s)
	next.labels = next.labels.Set(name, value)
	return next
}

// WithNoLabels removes all labels and returns a new signal.
func (s *Signal) WithNoLabels() *Signal {
	next := cloneSignal(s)
	next.labels = meta.NewLabels()
	return next
}

// WithoutLabels removes specific labels and returns a new signal.
func (s *Signal) WithoutLabels(names ...string) *Signal {
	next := cloneSignal(s)
	next.labels = next.labels.Remove(names...)
	return next
}

// Scalars returns a defensive copy of the signal's scalars store.
func (s *Signal) Scalars() *meta.Scalars {
	return cloneScalars(s.scalars)
}

// WithOnlyScalars replaces all scalars and returns a new signal.
func (s *Signal) WithOnlyScalars(scalarsMap map[string]float64) *Signal {
	next := cloneSignal(s)
	next.scalars = meta.NewScalars().SetMany(scalarsMap)
	return next
}

// WithScalars adds or updates scalars and returns a new signal.
func (s *Signal) WithScalars(scalarsMap map[string]float64) *Signal {
	next := cloneSignal(s)
	next.scalars = next.scalars.SetMany(scalarsMap)
	return next
}

// WithScalar adds or updates a single scalar and returns a new signal.
func (s *Signal) WithScalar(name string, value float64) *Signal {
	next := cloneSignal(s)
	next.scalars = next.scalars.Set(name, value)
	return next
}

// WithNoScalars removes all scalars and returns a new signal.
func (s *Signal) WithNoScalars() *Signal {
	next := cloneSignal(s)
	next.scalars = meta.NewScalars()
	return next
}

// WithoutScalars removes specific scalars and returns a new signal.
func (s *Signal) WithoutScalars(names ...string) *Signal {
	next := cloneSignal(s)
	next.scalars = next.scalars.Remove(names...)
	return next
}

// MapPayload applies a mapper function to the signal's payload and returns a new signal.
// The new signal preserves all labels and scalars from the original signal.
func (s *Signal) MapPayload(mapper PayloadMapper) *Signal {
	out := New(mapper(s.Payload()))
	out.labels = cloneLabels(s.labels)
	out.scalars = cloneScalars(s.scalars)
	return out
}

// Payload returns the signal's payload. The value is shallow: if the payload is
// a pointer, slice, or map, the caller must not mutate it — after fan-out the
// same payload may be shared by components activating concurrently.
//
// This does not fail. nil is a valid payload, and the only way to hold a signal
// without one is to have built a zero-value Signal instead of calling New —
// a construction bug, not a runtime condition, and not worth an error return on
// the single most-called accessor in the library. Such a signal reads as nil.
//
// For the payload as a concrete type, and an error when it is not that type,
// use As. For "was there a signal at all", check the group: Group.First returns
// nil for an empty one.
func (s *Signal) Payload() any {
	if len(s.payload) == 0 {
		return nil
	}
	return s.payload[0]
}

// Map applies a given mapper func and returns a new signal.
func (s *Signal) Map(m Mapper) *Signal {
	return m(cloneSignal(s))
}
