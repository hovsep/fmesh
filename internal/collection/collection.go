// Package collection holds the generic bases behind the group and collection
// types in signal, port, cycle and component.
package collection

import (
	"fmt"
	"iter"
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

// All returns an iterator over the items in insertion order. It allocates
// nothing; use slices.Collect(s.All()) for an independent slice. A loop body
// must not add to a mutating group it is ranging over.
func (s *Slice[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range s.items {
			if !yield(item) {
				return
			}
		}
	}
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

// The backing-slice accessors are package functions, not methods, so they are
// not promoted onto the embedding types' public surface (a mutator would break
// copy-on-write on signal.Group).

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
	byName map[string]T
	// ordered holds the same items sorted by name, for traversal. It is rebuilt
	// on membership change, never lazily: reads come from activation goroutines.
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

// Len returns the number of items.
func (k *Keyed[T]) Len() int {
	return len(k.byName)
}

// IsEmpty returns true when there are no items.
func (k *Keyed[T]) IsEmpty() bool {
	return k.Len() == 0
}

// All returns an iterator over the items in name order, the order every
// traversal on the collection uses. It allocates nothing; use
// slices.Collect(k.All()) for an independent slice.
//
// The iterator walks the sorted list as it was when ranging started: Add and
// Remove replace the list rather than change it, so a loop body may change the
// collection.
func (k *Keyed[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range k.ordered {
			if !yield(item) {
				return
			}
		}
	}
}

// First returns the first item in name order, or the zero value if empty.
func (k *Keyed[T]) First() T {
	if len(k.ordered) == 0 {
		var zero T
		return zero
	}
	return k.ordered[0]
}

// Any returns true if any item matches the predicate.
func (k *Keyed[T]) Any(pred func(T) bool) bool {
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

// Find returns the first item matching the predicate, in name order.
// Returns the zero value if no match found.
func (k *Keyed[T]) Find(pred func(T) bool) T {
	for _, item := range k.ordered {
		if pred(item) {
			return item
		}
	}
	var zero T
	return zero
}
