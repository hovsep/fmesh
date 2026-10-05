package signal

import (
	"github.com/hovsep/fmesh/meta"
)

// Signal is a wrapper around the data flowing between components.
// Mutating-style methods return a new *Signal; receivers are never modified.
type Signal struct {
	meta    *meta.Meta // nil until metadata is added
	payload any
}

// cloneSignal returns a copy of s with an independent metadata store, or none
// when s has none. Payload is shallow-copied.
func cloneSignal(s *Signal) *Signal {
	if s == nil {
		return nil
	}
	next := &Signal{payload: s.payload}
	if s.meta != nil {
		next.meta = s.meta.Clone()
	}
	return next
}

// cloneForMeta is cloneSignal for a caller about to write metadata: the copy
// always has a store.
func cloneForMeta(s *Signal) *Signal {
	return &Signal{
		meta:    s.meta.Clone(),
		payload: s.payload,
	}
}

// New creates a new signal with the given payload.
func New(payload any) *Signal {
	return &Signal{
		payload: payload,
	}
}

// Meta returns a copy of the signal's metadata. Mutating it changes nothing;
// use WithMeta and assign the result.
func (s *Signal) Meta() *meta.Meta {
	return s.meta.Clone()
}

// WithMeta adds or updates one metadata entry and returns a new signal.
func (s *Signal) WithMeta[T meta.Value](key string, value T) *Signal {
	next := cloneForMeta(s)
	next.meta.Set(key, value)
	return next
}

// WithMetaMany adds or updates every entry of values and returns a new signal.
func (s *Signal) WithMetaMany[T meta.Value](values map[string]T) *Signal {
	next := cloneForMeta(s)
	next.meta.SetMany(values)
	return next
}

// WithoutMeta removes the named metadata entries and returns a new signal.
func (s *Signal) WithoutMeta(keys ...string) *Signal {
	next := cloneForMeta(s)
	next.meta.Remove(keys...)
	return next
}

// MapPayload applies a mapper function to the signal's payload and returns a new signal.
// The new signal preserves all metadata from the original signal.
func (s *Signal) MapPayload(mapper PayloadMapper) *Signal {
	out := cloneSignal(s)
	out.payload = mapper(s.Payload())
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
// use [Signal.As]. For "was there a signal at all", check the group: Group.First returns
// nil for an empty one.
func (s *Signal) Payload() any {
	return s.payload
}

// Map applies a given mapper func and returns a new signal.
func (s *Signal) Map(m Mapper) *Signal {
	return m(cloneSignal(s))
}
