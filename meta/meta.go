package meta

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
)

// Value is the set of types a Meta entry can hold: string, bool, every Go
// integer and float type, and named types built on them, such as
// `type Celsius float64`. An entry reads back only as the type it was written
// with: Set("n", 3) stores an int, so read it with Value[int], not
// Value[int64] or Value[float64]; Set("t", Celsius(36.6)) reads back with
// Value[Celsius], not Value[float64].
type Value interface {
	cmp.Ordered | ~bool
}

// Predicate tests one entry. value holds one of the [Value] types.
type Predicate func(key string, value any) bool

// Meta is a mutable key→value store for string, bool and numeric metadata.
// Write methods modify the receiver in place and return it for chaining.
// Clone and Filter are the non-mutating methods; each returns a new Meta.
type Meta struct {
	entries map[string]any // nil until the first write; reads of a nil map are safe
}

// New creates an empty Meta.
func New() *Meta {
	return &Meta{}
}

// Set adds or updates one entry (upsert semantics).
func (m *Meta) Set[T Value](key string, value T) *Meta {
	if m.entries == nil {
		m.entries = make(map[string]any)
	}
	m.entries[key] = value
	return m
}

// SetMany adds or updates every entry of values (upsert semantics).
func (m *Meta) SetMany[T Value](values map[string]T) *Meta {
	if m.entries == nil && len(values) > 0 {
		m.entries = make(map[string]any, len(values))
	}
	for k, v := range values {
		m.entries[k] = v
	}
	return m
}

// Remove deletes the named entries. Missing keys are silently ignored.
func (m *Meta) Remove(keys ...string) *Meta {
	for _, k := range keys {
		delete(m.entries, k)
	}
	return m
}

// Clear removes every entry.
func (m *Meta) Clear() *Meta {
	clear(m.entries)
	return m
}

// Value returns the entry for key as T. It fails when the key is absent or
// holds another type; the error names the key either way.
func (m *Meta) Value[T Value](key string) (T, error) {
	var zero T
	raw, ok := m.entries[key]
	if !ok {
		return zero, fmt.Errorf("meta %q not found", key)
	}
	v, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("meta %q holds %T, not %T", key, raw, zero)
	}
	return v, nil
}

// ValueOrDefault returns the entry for key as T, or def when the key is absent
// or holds another type. T is inferred from def: 0 reads an int, 0.0 a float64.
func (m *Meta) ValueOrDefault[T Value](key string, def T) T {
	if v, ok := m.entries[key].(T); ok {
		return v
	}
	return def
}

// ValueIs reports whether key is present and holds exactly value.
func (m *Meta) ValueIs[T Value](key string, value T) bool {
	v, ok := m.entries[key].(T)
	return ok && v == value
}

// Has reports whether every key is present. No keys is vacuously true.
func (m *Meta) Has(keys ...string) bool {
	for _, k := range keys {
		if _, ok := m.entries[k]; !ok {
			return false
		}
	}
	return true
}

// HasAny reports whether at least one key is present.
func (m *Meta) HasAny(keys ...string) bool {
	for _, k := range keys {
		if _, ok := m.entries[k]; ok {
			return true
		}
	}
	return false
}

// Keys returns every key, sorted. The caller owns the slice.
func (m *Meta) Keys() []string {
	return slices.Sorted(maps.Keys(m.entries))
}

// Len returns the number of entries.
func (m *Meta) Len() int {
	return len(m.entries)
}

// IsEmpty reports whether the store holds nothing.
func (m *Meta) IsEmpty() bool {
	return len(m.entries) == 0
}

// All returns a copy of every entry. Mutating it does not affect the store.
func (m *Meta) All() map[string]any {
	return maps.Clone(m.entries)
}

// Clone returns an independent copy. A nil receiver clones to an empty Meta.
func (m *Meta) Clone() *Meta {
	c := New()
	if m != nil {
		c.entries = maps.Clone(m.entries)
	}
	return c
}

// Filter returns a new Meta holding the entries that pass pred.
func (m *Meta) Filter(pred Predicate) *Meta {
	filtered := New()
	for k, v := range m.entries {
		if pred(k, v) {
			if filtered.entries == nil {
				filtered.entries = make(map[string]any)
			}
			filtered.entries[k] = v
		}
	}
	return filtered
}
