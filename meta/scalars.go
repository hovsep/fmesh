package meta

import (
	"maps"
	"math"
	"slices"
)

// Scalars is a mutable name→float64 store for numeric metadata.
// All write methods modify the receiver in place.
type Scalars struct {
	scalars map[string]float64
}

// NewScalars creates an initialized, empty Scalars store.
func NewScalars() *Scalars {
	return &Scalars{
		scalars: make(map[string]float64),
	}
}

// All returns a defensive copy of all scalars. Mutating the returned map does
// not change the store.
func (s *Scalars) All() map[string]float64 {
	return maps.Clone(s.scalars)
}

// Keys returns all scalar names as a sorted slice. The caller owns the slice.
func (s *Scalars) Keys() []string {
	keys := make([]string, 0, len(s.scalars))
	for k := range s.scalars {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Set adds or updates a single scalar (upsert semantics).
func (s *Scalars) Set(name string, value float64) *Scalars {
	s.scalars[name] = value
	return s
}

// SetMany adds or updates multiple scalars (upsert semantics).
func (s *Scalars) SetMany(scalars map[string]float64) *Scalars {
	for name, value := range scalars {
		s.Set(name, value)
	}
	return s
}

// Get returns the value for name and true, or 0 and false if not present.
func (s *Scalars) Get(name string) (float64, bool) {
	v, ok := s.scalars[name]
	return v, ok
}

// GetOrDefault returns the value for name, or def if name is not present.
func (s *Scalars) GetOrDefault(name string, def float64) float64 {
	if v, ok := s.scalars[name]; ok {
		return v
	}
	return def
}

// Has returns true when the store contains name.
func (s *Scalars) Has(name string) bool {
	_, ok := s.scalars[name]
	return ok
}

// Remove deletes the named scalars. Missing names are silently ignored.
func (s *Scalars) Remove(names ...string) *Scalars {
	for _, name := range names {
		delete(s.scalars, name)
	}
	return s
}

// Clear removes all scalars.
func (s *Scalars) Clear() *Scalars {
	s.scalars = make(map[string]float64)
	return s
}

// Len returns the number of scalars.
func (s *Scalars) Len() int {
	return len(s.scalars)
}

// IsEmpty returns true when there are no scalars in the store.
func (s *Scalars) IsEmpty() bool {
	return s.Len() == 0
}

// Min returns the name and value of the entry with the smallest value.
// ok is false when the store is empty.
func (s *Scalars) Min() (name string, value float64, ok bool) {
	value = math.MaxFloat64
	for k, v := range s.scalars {
		if v < value || !ok {
			name, value, ok = k, v, true
		}
	}
	return
}

// Max returns the name and value of the entry with the largest value.
// ok is false when the store is empty.
func (s *Scalars) Max() (name string, value float64, ok bool) {
	value = -math.MaxFloat64
	for k, v := range s.scalars {
		if v > value || !ok {
			name, value, ok = k, v, true
		}
	}
	return
}

// Sum returns the sum of the given scalar names.
// If no names are given, it sums all scalars.
// Missing names contribute 0.
func (s *Scalars) Sum(names ...string) float64 {
	if len(names) == 0 {
		var total float64
		for _, v := range s.scalars {
			total += v
		}
		return total
	}
	var total float64
	for _, name := range names {
		total += s.scalars[name]
	}
	return total
}

// Average returns the mean of the given scalar names and true.
// If no names are given, it averages all scalars.
// ok is false when there are no values to average (empty store or no names given with empty store).
func (s *Scalars) Average(names ...string) (float64, bool) {
	if len(names) == 0 {
		if s.IsEmpty() {
			return 0, false
		}
		return s.Sum() / float64(len(s.scalars)), true
	}
	return s.Sum(names...) / float64(len(names)), true
}

// Scale multiplies the named scalar by factor in place. No-op if name is absent.
func (s *Scalars) Scale(name string, factor float64) *Scalars {
	if v, ok := s.scalars[name]; ok {
		s.scalars[name] = v * factor
	}
	return s
}

// Merge returns a new Scalars containing all entries from both s and other.
// On key conflict, other's value wins. Neither s nor other is modified.
func (s *Scalars) Merge(other *Scalars) *Scalars {
	merged := NewScalars()
	maps.Copy(merged.scalars, s.scalars)
	maps.Copy(merged.scalars, other.scalars)
	return merged
}

// Every returns true if all scalars satisfy the predicate.
// Returns true for an empty store (vacuous truth).
func (s *Scalars) Every(pred ScalarPredicate) bool {
	for k, v := range s.scalars {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one scalar satisfies the predicate.
func (s *Scalars) Any(pred ScalarPredicate) bool {
	for k, v := range s.scalars {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// Count returns the number of scalars that match the predicate.
func (s *Scalars) Count(pred ScalarPredicate) int {
	count := 0
	for k, v := range s.scalars {
		if pred(k, v) {
			count++
		}
	}
	return count
}

// Filter returns a new Scalars with entries that pass the predicate.
func (s *Scalars) Filter(pred ScalarPredicate) *Scalars {
	filtered := NewScalars()
	for k, v := range s.scalars {
		if pred(k, v) {
			filtered.Set(k, v)
		}
	}
	return filtered
}

// ForEach applies action to each scalar. Returns the first error encountered.
func (s *Scalars) ForEach(action func(name string, value float64) error) error {
	for k, v := range s.scalars {
		if err := action(k, v); err != nil {
			return err
		}
	}
	return nil
}
