// Package collection holds the generic bases behind the group and collection
// types in signal, port, cycle and component.
package collection

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Slice is the shared read surface for slice-backed groups. It preserves
// insertion order; whether the embedding type mutates it in place or treats it
// as copy-on-write is the embedding type's contract.
type Slice[T any] struct {
	items []T
}

// Len returns the number of items.
func (s *Slice[T]) Len() int {
	return len(s.items)
}

// IsEmpty returns true when there are no items.
func (s *Slice[T]) IsEmpty() bool {
	return s.Len() == 0
}

// All returns a cloned slice of items. The slice is independent of the
// collection; the items inside are shared.
func (s *Slice[T]) All() []T {
	return slices.Clone(s.items)
}

// First returns the first item, or the zero value if empty.
func (s *Slice[T]) First() T {
	if s.IsEmpty() {
		var zero T
		return zero
	}
	return s.items[0]
}

// Last returns the last item, or the zero value if empty.
func (s *Slice[T]) Last() T {
	if s.IsEmpty() {
		var zero T
		return zero
	}
	return s.items[len(s.items)-1]
}

// Any returns true if at least one item matches the predicate.
func (s *Slice[T]) Any(pred func(T) bool) bool {
	return slices.ContainsFunc(s.items, pred)
}

// Every returns true if all items match the predicate.
// Returns true when empty (vacuous truth).
func (s *Slice[T]) Every(pred func(T) bool) bool {
	for _, item := range s.items {
		if !pred(item) {
			return false
		}
	}
	return true
}

// Count returns the number of items that match the predicate.
func (s *Slice[T]) Count(pred func(T) bool) int {
	count := 0
	for _, item := range s.items {
		if pred(item) {
			count++
		}
	}
	return count
}

// Find returns the first item matching the predicate, or the zero value if none match.
func (s *Slice[T]) Find(pred func(T) bool) T {
	for _, item := range s.items {
		if pred(item) {
			return item
		}
	}
	var zero T
	return zero
}

// ForEach applies the action to each item. Returns the first error encountered.
func (s *Slice[T]) ForEach(action func(T) error) error {
	for _, item := range s.items {
		if err := action(item); err != nil {
			return err
		}
	}
	return nil
}

// ForEachIf applies the action only to items that match the predicate.
// Returns the first error encountered.
func (s *Slice[T]) ForEachIf(pred func(T) bool, action func(T) error) error {
	for _, item := range s.items {
		if pred(item) {
			if err := action(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// The backing-slice accessors are package functions, not methods: a method
// would be promoted onto the embedding types' public method sets, and on the
// copy-on-write types a public mutator reopens the hole Group.Meta() once
// was. A function in an internal package is unreachable from outside the
// module.

// Items returns the live backing slice, not a clone — internal hot paths only.
// Callers on copy-on-write types must not mutate it.
func Items[T any](s *Slice[T]) []T {
	return s.items
}

// SetItems swaps the backing slice for the given one.
func SetItems[T any](s *Slice[T], items []T) {
	s.items = items
}

// AppendItems appends items to the backing slice in place.
func AppendItems[T any](s *Slice[T], items ...T) {
	s.items = append(s.items, items...)
}

// Named is an element a Keyed collection can index.
type Named interface {
	Name() string
}

// Keyed is a name-indexed collection with deterministic name-ordered
// traversal. It cannot carry two items with the same name.
type Keyed[T Named] struct {
	// byName indexes by name, for ByName.
	byName map[string]T
	// ordered holds the same items, sorted by name — traversal reads this, so it
	// costs neither a per-element map lookup nor an allocation. It is updated when
	// membership changes rather than lazily on read: items are read from
	// activation goroutines, and a lazily filled cache would turn a read into a
	// write.
	ordered []T
	// kind is the element noun used in error messages, e.g. "port".
	kind string
}

// NewKeyed creates an empty collection whose error messages call elements kind.
func NewKeyed[T Named](kind string) *Keyed[T] {
	return &Keyed[T]{
		byName: make(map[string]T),
		kind:   kind,
	}
}

func byName[T Named](a, b T) int {
	return strings.Compare(a.Name(), b.Name())
}

// ByName returns the item with the given name, or the zero value if absent.
func (k *Keyed[T]) ByName(name string) T {
	return k.byName[name]
}

// Add adds items and returns an error on a name conflict, with the collection
// or within items. On error nothing is added.
//
// The new items are merged into the sorted list in one pass, so adding n items
// at once costs O(n log n), not a re-sort per item. The list is replaced, never
// changed in place: a traversal already ranging over it keeps its snapshot.
func (k *Keyed[T]) Add(items ...T) error {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		_, exists := k.byName[item.Name()]
		_, repeated := seen[item.Name()]
		if exists || repeated {
			return fmt.Errorf("%s %q already exists", k.kind, item.Name())
		}
		seen[item.Name()] = struct{}{}
	}
	for _, item := range items {
		k.byName[item.Name()] = item
	}

	added := slices.SortedFunc(slices.Values(items), byName[T])
	merged := make([]T, 0, len(k.ordered)+len(added))
	i, j := 0, 0
	for i < len(k.ordered) && j < len(added) {
		if byName(k.ordered[i], added[j]) < 0 {
			merged = append(merged, k.ordered[i])
			i++
		} else {
			merged = append(merged, added[j])
			j++
		}
	}
	merged = append(merged, k.ordered[i:]...)
	merged = append(merged, added[j:]...)
	k.ordered = merged
	return nil
}

// Remove deletes items by name. Names that match nothing are ignored.
func (k *Keyed[T]) Remove(names ...string) {
	for _, name := range names {
		delete(k.byName, name)
	}
	k.ordered = slices.DeleteFunc(slices.Clone(k.ordered), func(item T) bool {
		_, kept := k.byName[item.Name()]
		return !kept
	})
}

// Reset removes all items. A function for the same reason as Items: keeping
// Clear off the method set keeps it off every embedding type's public surface.
func Reset[T Named](k *Keyed[T]) {
	k.byName = make(map[string]T)
	k.ordered = nil
}

// Len returns the number of items.
func (k *Keyed[T]) Len() int {
	return len(k.byName)
}

// IsEmpty returns true when there are no items.
func (k *Keyed[T]) IsEmpty() bool {
	return k.Len() == 0
}

// All returns a shallow copy of all items as a map.
// A copy is returned so the caller cannot mutate the internal state.
func (k *Keyed[T]) All() map[string]T {
	return maps.Clone(k.byName)
}

// AllOrdered returns all items sorted by name.
// This is the order every traversal on the collection uses.
// A copy is returned so the caller cannot reorder the collection's own state;
// internal traversals range over Each instead.
func (k *Keyed[T]) AllOrdered() []T {
	return slices.Clone(k.ordered)
}

// Each yields every item in name order, without allocating and without a map
// lookup per item.
//
// This is the internal form of AllOrdered, for traversals that run for every
// component on every cycle, where neither a slice copy nor a hash lookup per
// item belongs. Range over it: for item := range k.Each.
func (k *Keyed[T]) Each(yield func(T) bool) {
	for _, item := range k.ordered {
		if !yield(item) {
			return
		}
	}
}

// AnyMatch returns true if any item matches the predicate.
func (k *Keyed[T]) AnyMatch(pred func(T) bool) bool {
	return slices.ContainsFunc(k.ordered, pred)
}

// Every returns true if all items match the predicate.
// Returns true when empty (vacuous truth).
func (k *Keyed[T]) Every(pred func(T) bool) bool {
	for _, item := range k.ordered {
		if !pred(item) {
			return false
		}
	}
	return true
}

// Count returns the number of items that match the predicate.
func (k *Keyed[T]) Count(pred func(T) bool) int {
	count := 0
	for _, item := range k.ordered {
		if pred(item) {
			count++
		}
	}
	return count
}

// FindAny returns the first item matching the predicate, in name order.
// Returns the zero value if no match found.
func (k *Keyed[T]) FindAny(pred func(T) bool) T {
	for _, item := range k.ordered {
		if pred(item) {
			return item
		}
	}
	var zero T
	return zero
}

// ForEach applies the action to each item in name order.
// Returns the first error encountered.
func (k *Keyed[T]) ForEach(action func(T) error) error {
	for _, item := range k.ordered {
		if err := action(item); err != nil {
			return err
		}
	}
	return nil
}
