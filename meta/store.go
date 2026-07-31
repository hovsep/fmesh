package meta

import (
	"maps"
	"slices"
)

// store is the map[string]T behavior shared by Labels and Scalars.
//
// S is the embedding type, carried so mutating methods can return it and keep
// the stores chainable: without it, an embedded Set would return *store and
// c.Labels().Clear().SetMany(m) would not compile. It is the only per-instance
// cost of sharing, so nothing else lives here — Value is implemented by each
// concrete type rather than storing a name to put in its error message.
type store[T comparable, S any] struct {
	entries map[string]T
	self    S
}

func (s *store[T, S]) init(self S) {
	s.entries = make(map[string]T)
	s.self = self
}

// All returns a defensive copy of every entry. Mutating it does not affect the store.
func (s *store[T, S]) All() map[string]T {
	return maps.Clone(s.entries)
}

// Keys returns all names as a sorted slice. The caller owns the slice.
func (s *store[T, S]) Keys() []string {
	return slices.Sorted(maps.Keys(s.entries))
}

// Set adds or updates a single entry (upsert semantics).
func (s *store[T, S]) Set(name string, value T) S {
	s.entries[name] = value
	return s.self
}

// SetMany adds or updates multiple entries (upsert semantics).
func (s *store[T, S]) SetMany(entries map[string]T) S {
	maps.Copy(s.entries, entries)
	return s.self
}

// lookup returns the value for name and whether it was present.
func (s *store[T, S]) lookup(name string) (T, bool) {
	v, ok := s.entries[name]
	return v, ok
}

// ValueOrDefault returns the value for name, or def when it is absent.
func (s *store[T, S]) ValueOrDefault(name string, def T) T {
	if v, ok := s.entries[name]; ok {
		return v
	}
	return def
}

// ValueIs returns true when name is present and holds value.
func (s *store[T, S]) ValueIs(name string, value T) bool {
	v, ok := s.entries[name]
	return ok && v == value
}

// Has returns true when the store contains name.
func (s *store[T, S]) Has(name string) bool {
	_, ok := s.entries[name]
	return ok
}

// Remove deletes the named entries. Missing names are silently ignored.
func (s *store[T, S]) Remove(names ...string) S {
	for _, name := range names {
		delete(s.entries, name)
	}
	return s.self
}

// Clear removes every entry.
func (s *store[T, S]) Clear() S {
	clear(s.entries)
	return s.self
}

// Len returns the number of entries.
func (s *store[T, S]) Len() int {
	return len(s.entries)
}

// IsEmpty returns true when the store holds nothing.
func (s *store[T, S]) IsEmpty() bool {
	return s.Len() == 0
}

// Every returns true if all entries satisfy the predicate.
// An empty store returns true (vacuous truth).
func (s *store[T, S]) Every(pred func(string, T) bool) bool {
	for k, v := range s.entries {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one entry satisfies the predicate.
func (s *store[T, S]) Any(pred func(string, T) bool) bool {
	for k, v := range s.entries {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// Count returns the number of entries matching the predicate.
func (s *store[T, S]) Count(pred func(string, T) bool) int {
	count := 0
	for k, v := range s.entries {
		if pred(k, v) {
			count++
		}
	}
	return count
}

// ForEach applies action to each entry. Returns the first error encountered.
func (s *store[T, S]) ForEach(action func(string, T) error) error {
	for k, v := range s.entries {
		if err := action(k, v); err != nil {
			return err
		}
	}
	return nil
}

// filterInto copies the entries matching pred into dst.
// Filter stays on the concrete types because only they can build their own zero value.
func (s *store[T, S]) filterInto(dst map[string]T, pred func(string, T) bool) {
	for k, v := range s.entries {
		if pred(k, v) {
			dst[k] = v
		}
	}
}

// mergeInto copies s's entries then other's into dst, so other wins on conflict.
func (s *store[T, S]) mergeInto(dst, other map[string]T) {
	maps.Copy(dst, s.entries)
	maps.Copy(dst, other)
}
