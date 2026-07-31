package meta

import (
	"maps"
	"slices"
)

// store is the map[string]T behavior shared by Labels and Scalars.
//
// It holds the map and nothing else, so a Labels is exactly the size of the map
// header it wraps — an earlier version also carried a pointer back to the
// embedding type so that promoted mutators could return it, and that doubled
// every store from 8 bytes to 16. Signals own two apiece, which showed up as ~6%
// more bytes per mesh run.
//
// The price is that the four chainable mutators are declared on Labels and
// Scalars rather than promoted from here: they need a concrete return type. Read
// methods are promoted, because their return types do not name the receiver.
type store[T comparable] struct {
	entries map[string]T
}

func (s *store[T]) init() {
	s.entries = make(map[string]T)
}

// All returns a defensive copy of every entry. Mutating it does not affect the store.
func (s *store[T]) All() map[string]T {
	return maps.Clone(s.entries)
}

// Keys returns all names as a sorted slice. The caller owns the slice.
func (s *store[T]) Keys() []string {
	return slices.Sorted(maps.Keys(s.entries))
}

// ValueOrDefault returns the value for name, or def when it is absent.
func (s *store[T]) ValueOrDefault(name string, def T) T {
	if v, ok := s.entries[name]; ok {
		return v
	}
	return def
}

// ValueIs returns true when name is present and holds value.
func (s *store[T]) ValueIs(name string, value T) bool {
	v, ok := s.entries[name]
	return ok && v == value
}

// Has returns true when the store contains name.
func (s *store[T]) Has(name string) bool {
	_, ok := s.entries[name]
	return ok
}

// Len returns the number of entries.
func (s *store[T]) Len() int {
	return len(s.entries)
}

// IsEmpty returns true when the store holds nothing.
func (s *store[T]) IsEmpty() bool {
	return s.Len() == 0
}

// Every returns true if all entries satisfy the predicate.
// An empty store returns true (vacuous truth).
func (s *store[T]) Every(pred func(string, T) bool) bool {
	for k, v := range s.entries {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one entry satisfies the predicate.
func (s *store[T]) Any(pred func(string, T) bool) bool {
	for k, v := range s.entries {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// Count returns the number of entries matching the predicate.
func (s *store[T]) Count(pred func(string, T) bool) int {
	count := 0
	for k, v := range s.entries {
		if pred(k, v) {
			count++
		}
	}
	return count
}

// ForEach applies action to each entry. Returns the first error encountered.
func (s *store[T]) ForEach(action func(string, T) error) error {
	for k, v := range s.entries {
		if err := action(k, v); err != nil {
			return err
		}
	}
	return nil
}

// lookup returns the value for name and whether it was present. Value is
// implemented by each concrete type so its error can name what was missing.
func (s *store[T]) lookup(name string) (T, bool) {
	v, ok := s.entries[name]
	return v, ok
}

// The mutators below are unexported; Labels and Scalars wrap each one so the
// chain returns the concrete type.

func (s *store[T]) set(name string, value T) {
	s.entries[name] = value
}

func (s *store[T]) setMany(entries map[string]T) {
	maps.Copy(s.entries, entries)
}

func (s *store[T]) remove(names ...string) {
	for _, name := range names {
		delete(s.entries, name)
	}
}

func (s *store[T]) clear() {
	clear(s.entries)
}

// filterInto copies the entries matching pred into dst.
func (s *store[T]) filterInto(dst map[string]T, pred func(string, T) bool) {
	for k, v := range s.entries {
		if pred(k, v) {
			dst[k] = v
		}
	}
}

// mergeInto copies s's entries then other's into dst, so other wins on conflict.
func (s *store[T]) mergeInto(dst, other map[string]T) {
	maps.Copy(dst, s.entries)
	maps.Copy(dst, other)
}
