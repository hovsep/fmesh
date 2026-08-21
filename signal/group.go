package signal

import (
	"slices"

	"github.com/hovsep/fmesh/internal/collection"
	"github.com/hovsep/fmesh/meta"
)

// Group represents an ordered list of signals.
// Like Signal it is copy-on-write: the embedded slice is never mutated after
// construction; every mutator builds a new group.
// signalSlice hides the embedded field name so it cannot be reached or
// reassigned from outside; only the base's exported read methods promote.
type signalSlice = collection.Slice[*Signal]

// Group represents an ordered list of signals.
type Group struct {
	signalSlice
	labels  *meta.Labels
	scalars *meta.Scalars
}

func newGroupFromSignals(signals []*Signal) *Group {
	g := &Group{
		labels:  meta.NewLabels(),
		scalars: meta.NewScalars(),
	}
	g.replace(slices.Clone(signals))
	return g
}

// NewGroup creates a new group from the given payloads.
func NewGroup(payloads ...any) *Group {
	signals := make([]*Signal, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroupFromSignals(signals)
}

// raw and replace are the only touchpoints with the backing slice; Group is
// copy-on-write, so raw's result is never mutated.
func (g *Group) raw() []*Signal { return collection.Items(&g.signalSlice) }

func (g *Group) replace(items []*Signal) { collection.SetItems(&g.signalSlice, items) }

// Labels returns a copy of the group's own labels. Group is copy-on-write:
// mutate via WithLabel, which returns a new group — matching Signal.Labels.
func (g *Group) Labels() *meta.Labels { return cloneLabels(g.labels) }

// WithLabel adds or updates a single label on the group and returns a new group.
func (g *Group) WithLabel(name, value string) *Group {
	next := newGroupFromSignals(g.raw())
	next.labels = cloneLabels(g.labels)
	next.scalars = cloneScalars(g.scalars)
	next.labels.Set(name, value)
	return next
}

// Scalars returns a copy of the group's own scalars. Group is copy-on-write:
// mutate via WithScalar, which returns a new group — matching Signal.Scalars.
func (g *Group) Scalars() *meta.Scalars { return cloneScalars(g.scalars) }

// WithScalar adds or updates a single scalar on the group and returns a new group.
func (g *Group) WithScalar(name string, value float64) *Group {
	next := newGroupFromSignals(g.raw())
	next.labels = cloneLabels(g.labels)
	next.scalars = cloneScalars(g.scalars)
	next.scalars.Set(name, value)
	return next
}

// copyGroupMeta copies the group's own labels and scalars into dst.
func copyGroupMeta(src, dst *Group) *Group {
	dst.labels = cloneLabels(src.labels)
	dst.scalars = cloneScalars(src.scalars)
	return dst
}

// WithLabelOnEach returns a new group with each signal having the label set.
// The group's own labels and scalars are preserved on the returned group.
func (g *Group) WithLabelOnEach(name, value string) *Group {
	return copyGroupMeta(g, g.Map(func(s *Signal) *Signal {
		return s.WithLabel(name, value)
	}))
}

// WithScalarOnEach returns a new group with each signal having the scalar set.
// The group's own labels and scalars are preserved on the returned group.
func (g *Group) WithScalarOnEach(name string, value float64) *Group {
	return copyGroupMeta(g, g.Map(func(s *Signal) *Signal {
		return s.WithScalar(name, value)
	}))
}

// RemoveLabelOnEach returns a new group with each signal having the label removed.
// The group's own labels and scalars are preserved on the returned group.
func (g *Group) RemoveLabelOnEach(names ...string) *Group {
	return copyGroupMeta(g, g.Map(func(s *Signal) *Signal {
		return s.WithoutLabels(names...)
	}))
}

// RemoveScalarOnEach returns a new group with each signal having the scalar removed.
// The group's own labels and scalars are preserved on the returned group.
func (g *Group) RemoveScalarOnEach(names ...string) *Group {
	return copyGroupMeta(g, g.Map(func(s *Signal) *Signal {
		return s.WithoutScalars(names...)
	}))
}

// Contains returns true if the group contains the exact signal (pointer identity).
func (g *Group) Contains(s *Signal) bool {
	return slices.Contains(g.raw(), s)
}

// ContainsPayload reports whether any signal's payload equals the given value.
// T must be comparable, so a slice or map argument does not compile — use
// ContainsPayloadFunc for those. A nil payload is found with ContainsPayload[any](nil).
func (g *Group) ContainsPayload[T comparable](payload T) bool {
	target := any(payload)
	return g.ContainsPayloadFunc(func(p any) bool { return p == target })
}

// ContainsPayloadFunc returns true if any signal's payload satisfies eq.
func (g *Group) ContainsPayloadFunc(eq func(payload any) bool) bool {
	for _, sig := range g.raw() {
		if eq(sig.Payload()) {
			return true
		}
	}
	return false
}

// FirstPayload returns the payload of the first signal with error handling.
func (g *Group) FirstPayload() (any, error) {
	first := g.First()
	if first == nil {
		return nil, ErrNoSignalsInGroup
	}
	return first.Payload(), nil
}

// FirstPayloadOrDefault returns the first signal's payload as T, or the default
// when the group is empty or the payload is missing or of another type. T is
// inferred from the default; FirstPayloadOrDefault[any](x) accepts any payload.
func (g *Group) FirstPayloadOrDefault[T any](defaultValue T) T {
	return g.First().PayloadOrDefault(defaultValue)
}

// FirstPayloadOrNil returns the payload of the first signal or nil.
func (g *Group) FirstPayloadOrNil() any {
	return g.FirstPayloadOrDefault[any](nil)
}

// FirstAs returns the first signal's payload as T. An empty group is
// ErrNoSignalsInGroup; a payload of another type is the error As reports.
func (g *Group) FirstAs[T any]() (T, error) {
	first := g.First()
	if first == nil {
		var zero T
		return zero, ErrNoSignalsInGroup
	}
	return first.As[T]()
}

// AllPayloads returns a slice with all payloads of all signals in the group.
func (g *Group) AllPayloads() []any {
	all := make([]any, g.Len())
	for i, sig := range g.raw() {
		all[i] = sig.Payload()
	}
	return all
}

// With returns a new group with the given signals appended. The receiver is never modified.
// Nil signals are silently skipped.
func (g *Group) With(signals ...*Signal) *Group {
	newSignals := make([]*Signal, 0, g.Len()+len(signals))
	newSignals = append(newSignals, g.raw()...)
	for _, sig := range signals {
		if sig == nil {
			continue
		}
		newSignals = append(newSignals, sig)
	}
	return newGroupFromSignals(newSignals)
}

// WithPayloads returns a new group with signals created from the given payloads appended.
func (g *Group) WithPayloads(payloads ...any) *Group {
	newSignals := make([]*Signal, g.Len()+len(payloads))
	copy(newSignals, g.raw())
	for i, p := range payloads {
		newSignals[g.Len()+i] = New(p)
	}
	return newGroupFromSignals(newSignals)
}

// Join returns a new group containing signals from both groups. A nil other is
// treated as empty, matching With's treatment of nil signals.
func (g *Group) Join(other *Group) *Group {
	if other == nil {
		return newGroupFromSignals(g.raw())
	}
	newSignals := make([]*Signal, g.Len()+other.Len())
	copy(newSignals, g.raw())
	copy(newSignals[g.Len():], other.raw())
	return newGroupFromSignals(newSignals)
}

// Filter returns a new group with signals that pass the predicate.
func (g *Group) Filter(p Predicate) *Group {
	filtered := make([]*Signal, 0, g.Len())
	for _, s := range g.raw() {
		if p(s) {
			filtered = append(filtered, s)
		}
	}
	return newGroupFromSignals(filtered)
}

// Map returns a new group with every signal transformed by the mapper.
// Nil mapper results are dropped.
func (g *Group) Map(m Mapper) *Group {
	mapped := make([]*Signal, 0, g.Len())
	for _, s := range g.raw() {
		if result := m(cloneSignal(s)); result != nil {
			mapped = append(mapped, result)
		}
	}
	return newGroupFromSignals(mapped)
}

// MapIf is like Map but applies the mapper only to signals matching the predicate.
// Nil mapper results are dropped.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	mapped := make([]*Signal, 0, g.Len())
	for _, s := range g.raw() {
		cloned := cloneSignal(s)
		if predicate(s) {
			cloned = mapper(cloned)
		}
		if cloned != nil {
			mapped = append(mapped, cloned)
		}
	}
	return newGroupFromSignals(mapped)
}

// MapPayloads returns a new group with every payload transformed by the mapper.
func (g *Group) MapPayloads(mapper PayloadMapper) *Group {
	mapped := make([]*Signal, 0, g.Len())
	for _, s := range g.raw() {
		mapped = append(mapped, s.MapPayload(mapper))
	}
	return newGroupFromSignals(mapped)
}

// MapPayloadsIf is like MapPayloads but applies the mapper only to signals matching the predicate.
func (g *Group) MapPayloadsIf(predicate Predicate, mapper PayloadMapper) *Group {
	mapped := make([]*Signal, g.Len())
	for i, s := range g.raw() {
		if predicate(s) {
			mapped[i] = s.MapPayload(mapper)
		} else {
			mapped[i] = cloneSignal(s)
		}
	}
	return newGroupFromSignals(mapped)
}

// Reduce accumulates all signals into a single signal using the given function.
func (g *Group) Reduce(initial *Signal, fn Reducer) *Signal {
	acc := initial
	for _, s := range g.raw() {
		acc = fn(acc, s)
	}
	return acc
}

// ReducePayloads folds every payload into an accumulator of type A.
func (g *Group) ReducePayloads[A any](initial A, fn func(acc A, payload any) A) A {
	acc := initial
	for _, s := range g.raw() {
		acc = fn(acc, s.Payload())
	}
	return acc
}
