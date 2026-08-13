package meta

import (
	"fmt"
)

// Scalars is a mutable name→float64 store for numeric metadata.
// All write methods modify the receiver in place.
type Scalars struct {
	store[float64]
}

// NewScalars creates an initialized, empty Scalars store.
func NewScalars() *Scalars {
	s := &Scalars{}
	s.init()
	return s
}

// Set adds or updates a single scalar (upsert semantics).
func (s *Scalars) Set(name string, value float64) *Scalars {
	s.set(name, value)
	return s
}

// SetMany adds or updates multiple scalars (upsert semantics).
func (s *Scalars) SetMany(scalars map[string]float64) *Scalars {
	s.setMany(scalars)
	return s
}

// Remove deletes the named scalars. Missing names are silently ignored.
func (s *Scalars) Remove(names ...string) *Scalars {
	s.remove(names...)
	return s
}

// Clear removes every scalar.
func (s *Scalars) Clear() *Scalars {
	s.clear()
	return s
}

// Value returns the value for name, or an error if not found.
func (s *Scalars) Value(name string) (float64, error) {
	v, ok := s.lookup(name)
	if !ok {
		return 0, fmt.Errorf("scalar %s not found", name)
	}
	return v, nil
}

// Merge returns a new Scalars containing all entries from both s and other.
// On key conflict, other's value wins. Neither s nor other is modified.
func (s *Scalars) Merge(other *Scalars) *Scalars {
	merged := NewScalars()
	s.mergeInto(merged.entries, other.entries)
	return merged
}

// Filter returns a new Scalars with entries that pass the predicate.
func (s *Scalars) Filter(pred ScalarPredicate) *Scalars {
	filtered := NewScalars()
	s.filterInto(filtered.entries, pred)
	return filtered
}
