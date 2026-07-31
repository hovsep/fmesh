package meta

import (
	"fmt"
	"math"
)

// Scalars is a mutable name→float64 store for numeric metadata.
// All write methods modify the receiver in place.
type Scalars struct {
	store[float64, *Scalars]
}

// NewScalars creates an initialized, empty Scalars store.
func NewScalars() *Scalars {
	s := &Scalars{}
	s.init(s)
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

// Min returns the name and value of the entry with the smallest value.
// ok is false when the store is empty.
func (s *Scalars) Min() (name string, value float64, ok bool) {
	value = math.MaxFloat64
	for k, v := range s.entries {
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
	for k, v := range s.entries {
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
	var total float64
	if len(names) == 0 {
		for _, v := range s.entries {
			total += v
		}
		return total
	}
	for _, name := range names {
		total += s.entries[name]
	}
	return total
}

// Average returns the mean of the given scalar names and true.
// If no names are given, it averages all scalars.
// ok is false when there are no values to average.
func (s *Scalars) Average(names ...string) (float64, bool) {
	if len(names) == 0 {
		if s.IsEmpty() {
			return 0, false
		}
		return s.Sum() / float64(s.Len()), true
	}
	return s.Sum(names...) / float64(len(names)), true
}

// Scale multiplies the named scalar by factor in place. No-op if name is absent.
func (s *Scalars) Scale(name string, factor float64) *Scalars {
	if v, ok := s.entries[name]; ok {
		s.entries[name] = v * factor
	}
	return s
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
