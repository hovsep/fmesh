package signal

import (
	"reflect"
	"slices"

	"github.com/hovsep/fmesh/meta"
)

// Group represents an ordered list of signals.
type Group struct {
	signals []*Signal
	labels  *meta.Labels
	scalars *meta.Scalars
}

func newGroupFromSignals(signals []*Signal) *Group {
	return &Group{
		signals: slices.Clone(signals),
		labels:  meta.NewLabels(),
		scalars: meta.NewScalars(),
	}
}

// NewGroup creates a new group from the given payloads.
func NewGroup(payloads ...any) *Group {
	signals := make([]*Signal, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroupFromSignals(signals)
}

// Labels returns the group's own labels store.
func (g *Group) Labels() *meta.Labels { return g.labels }

// WithLabel adds or updates a single label on the group and returns a new group.
func (g *Group) WithLabel(name, value string) *Group {
	next := newGroupFromSignals(g.signals)
	next.labels = cloneLabels(g.labels)
	next.scalars = cloneScalars(g.scalars)
	next.labels.Set(name, value)
	return next
}

// Scalars returns the group's own scalars store.
func (g *Group) Scalars() *meta.Scalars { return g.scalars }

// WithScalar adds or updates a single scalar on the group and returns a new group.
func (g *Group) WithScalar(name string, value float64) *Group {
	next := newGroupFromSignals(g.signals)
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

// First returns the first signal in the group, or nil if empty.
func (g *Group) First() *Signal {
	if g.IsEmpty() {
		return nil
	}
	return g.signals[0]
}

// Last returns the last signal in the group, or nil if empty.
func (g *Group) Last() *Signal {
	if g.IsEmpty() {
		return nil
	}
	return g.signals[len(g.signals)-1]
}

// IsEmpty returns true when there are no signals in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// Find returns the first signal matching the predicate, or nil if none match.
func (g *Group) Find(predicate Predicate) *Signal {
	if g.IsEmpty() {
		return nil
	}
	for _, s := range g.signals {
		if predicate(s) {
			return s
		}
	}
	return nil
}

// Any returns true if at least one signal matches the predicate.
func (g *Group) Any(p Predicate) bool {
	return slices.ContainsFunc(g.signals, p)
}

// Every returns true if all signals match the predicate.
// Returns true for an empty group (vacuous truth).
func (g *Group) Every(p Predicate) bool {
	for _, sig := range g.signals {
		if !p(sig) {
			return false
		}
	}
	return true
}

// Count returns the number of signals that match the predicate.
func (g *Group) Count(predicate Predicate) int {
	n := 0
	for _, sig := range g.signals {
		if predicate(sig) {
			n++
		}
	}
	return n
}

// Contains returns true if the group contains the exact signal (pointer identity).
func (g *Group) Contains(s *Signal) bool {
	return slices.Contains(g.signals, s)
}

// ContainsPayload returns true if any signal's payload equals the given value.
// Returns an error if the payload type is not comparable; use ContainsPayloadFunc for non-comparable types.
func (g *Group) ContainsPayload(payload any) (bool, error) {
	if payload != nil && !reflect.TypeOf(payload).Comparable() {
		return false, ErrPayloadNotComparable
	}
	return g.ContainsPayloadFunc(func(p any) bool {
		return p == payload
	}), nil
}

// ContainsPayloadFunc returns true if any signal's payload satisfies eq.
func (g *Group) ContainsPayloadFunc(eq func(payload any) bool) bool {
	for _, sig := range g.signals {
		p, err := sig.Payload()
		if err != nil {
			continue
		}
		if eq(p) {
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
	return first.Payload()
}

// FirstPayloadOrDefault returns the payload of the first signal or a default value.
func (g *Group) FirstPayloadOrDefault(defaultPayload any) any {
	payload, err := g.FirstPayload()
	if err != nil {
		return defaultPayload
	}
	return payload
}

// FirstPayloadOrNil returns the payload of the first signal or nil.
func (g *Group) FirstPayloadOrNil() any {
	return g.FirstPayloadOrDefault(nil)
}

// AllPayloads returns a slice with all payloads of all signals in the group.
func (g *Group) AllPayloads() ([]any, error) {
	all := make([]any, g.Len())
	for i, sig := range g.signals {
		p, err := sig.Payload()
		if err != nil {
			return nil, err
		}
		all[i] = p
	}
	return all, nil
}

// With returns a new group with the given signals appended. The receiver is never modified.
// Nil signals are silently skipped.
func (g *Group) With(signals ...*Signal) *Group {
	newSignals := make([]*Signal, 0, g.Len()+len(signals))
	newSignals = append(newSignals, g.signals...)
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
	copy(newSignals, g.signals)
	for i, p := range payloads {
		newSignals[g.Len()+i] = New(p)
	}
	return newGroupFromSignals(newSignals)
}

// Join returns a new group containing signals from both groups.
func (g *Group) Join(other *Group) *Group {
	newSignals := make([]*Signal, g.Len()+other.Len())
	copy(newSignals, g.signals)
	copy(newSignals[g.Len():], other.signals)
	return newGroupFromSignals(newSignals)
}

// All returns a cloned slice of signals. The slice is independent of the group;
// the *Signal pointers inside are shared, but Signal is copy-on-write so callers
// cannot corrupt group state through the returned pointers.
func (g *Group) All() []*Signal {
	return slices.Clone(g.signals)
}

// Len returns the number of signals in the group.
func (g *Group) Len() int {
	return len(g.signals)
}

// ForEach applies the action to each signal. Returns the first error encountered (if any).
func (g *Group) ForEach(action func(*Signal) error) error {
	for _, s := range g.signals {
		if err := action(s); err != nil {
			return err
		}
	}
	return nil
}

// ForEachIf applies the action only to signals that match the predicate.
func (g *Group) ForEachIf(predicate Predicate, action func(*Signal) error) error {
	for _, s := range g.signals {
		if predicate(s) {
			if err := action(s); err != nil {
				return err
			}
		}
	}
	return nil
}

// Filter returns a new group with signals that pass the predicate.
func (g *Group) Filter(p Predicate) *Group {
	filtered := make([]*Signal, 0, len(g.signals))
	for _, s := range g.signals {
		if p(s) {
			filtered = append(filtered, s)
		}
	}
	return newGroupFromSignals(filtered)
}

// Map returns a new group with every signal transformed by the mapper.
func (g *Group) Map(m Mapper) *Group {
	mapped := make([]*Signal, 0, len(g.signals))
	for _, s := range g.signals {
		mapped = append(mapped, m(cloneSignal(s)))
	}
	return newGroupFromSignals(mapped)
}

// MapIf is like Map but applies the mapper only to signals matching the predicate.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	mapped := make([]*Signal, len(g.signals))
	for i, s := range g.signals {
		cloned := cloneSignal(s)
		if predicate(s) {
			mapped[i] = mapper(cloned)
		} else {
			mapped[i] = cloned
		}
	}
	return newGroupFromSignals(mapped)
}

// MapPayloads returns a new group with every payload transformed by the mapper.
func (g *Group) MapPayloads(mapper PayloadMapper) *Group {
	mapped := make([]*Signal, 0, len(g.signals))
	for _, s := range g.signals {
		mapped = append(mapped, s.MapPayload(mapper))
	}
	return newGroupFromSignals(mapped)
}

// MapPayloadsIf is like MapPayloads but applies the mapper only to signals matching the predicate.
func (g *Group) MapPayloadsIf(predicate Predicate, mapper PayloadMapper) *Group {
	mapped := make([]*Signal, len(g.signals))
	for i, s := range g.signals {
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
	for _, s := range g.signals {
		acc = fn(acc, s)
	}
	return acc
}

// ReducePayloads accumulates all signal payloads into a single value using the given function.
func (g *Group) ReducePayloads(initial any, fn PayloadReducer) any {
	acc := initial
	for _, s := range g.signals {
		p, err := s.Payload()
		if err != nil {
			continue
		}
		acc = fn(acc, p)
	}
	return acc
}

// SumScalar returns the sum of the named scalar across all signals in the group.
// Signals that do not have the scalar contribute 0. Returns 0 for an empty group.
func (g *Group) SumScalar(name string) float64 {
	var total float64
	for _, s := range g.signals {
		v, ok := s.scalars.Get(name)
		if ok {
			total += v
		}
	}
	return total
}

// MinScalar returns the minimum value of the named scalar across all signals.
// ok is false when no signal in the group has that scalar.
func (g *Group) MinScalar(name string) (float64, bool) {
	found := false
	var minVal float64
	for _, s := range g.signals {
		v, ok := s.scalars.Get(name)
		if !ok {
			continue
		}
		if !found || v < minVal {
			minVal, found = v, true
		}
	}
	return minVal, found
}

// MaxScalar returns the maximum value of the named scalar across all signals.
// ok is false when no signal in the group has that scalar.
func (g *Group) MaxScalar(name string) (float64, bool) {
	found := false
	var maxVal float64
	for _, s := range g.signals {
		v, ok := s.scalars.Get(name)
		if !ok {
			continue
		}
		if !found || v > maxVal {
			maxVal, found = v, true
		}
	}
	return maxVal, found
}

// AvgScalar returns the mean value of the named scalar across all signals that have it.
// ok is false when no signal in the group has that scalar.
func (g *Group) AvgScalar(name string) (float64, bool) {
	var sum float64
	count := 0
	for _, s := range g.signals {
		v, ok := s.scalars.Get(name)
		if !ok {
			continue
		}
		sum += v
		count++
	}
	if count == 0 {
		return 0, false
	}
	return sum / float64(count), true
}
