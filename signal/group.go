package signal

import (
	"reflect"
	"slices"
)

// Group represents an ordered list of signals.
type Group struct {
	chainableErr error
	signals      Signals
}

func newGroupFromSignals(signals Signals) *Group {
	return &Group{
		chainableErr: nil,
		signals:      slices.Clone(signals),
	}
}

// NewGroup creates a new group from the given payloads.
func NewGroup(payloads ...any) *Group {
	signals := make(Signals, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroupFromSignals(signals)
}

// First returns the first signal in the group, or nil if empty or errored.
func (g *Group) First() *Signal {
	if g.HasChainableErr() {
		return nil
	}
	if g.IsEmpty() {
		return nil
	}
	return g.signals[0]
}

// Last returns the last signal in the group, or nil if empty or errored.
func (g *Group) Last() *Signal {
	if g.HasChainableErr() {
		return nil
	}
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
	if g.HasChainableErr() || g.IsEmpty() {
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
	if g.HasChainableErr() || g.IsEmpty() {
		return false
	}
	return slices.ContainsFunc(g.signals, p)
}

// Every returns true if all signals match the predicate.
// Returns true for an empty group (vacuous truth).
func (g *Group) Every(p Predicate) bool {
	if g.HasChainableErr() {
		return false
	}
	for _, sig := range g.signals {
		if !p(sig) {
			return false
		}
	}
	return true
}

// Count returns the number of signals that match the predicate.
func (g *Group) Count(predicate Predicate) int {
	if g.HasChainableErr() {
		return 0
	}
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
	if g.HasChainableErr() || g.IsEmpty() {
		return false
	}
	return slices.Contains(g.signals, s)
}

// ContainsPayload returns true if any signal's payload equals the given value.
// Panics if the payload type is not comparable; use ContainsPayloadFunc instead.
func (g *Group) ContainsPayload(payload any) bool {
	if payload != nil && !reflect.TypeOf(payload).Comparable() {
		panic("ContainsPayload: payload type is not comparable, use ContainsPayloadFunc instead")
	}
	return g.ContainsPayloadFunc(func(p any) bool {
		return p == payload
	})
}

// ContainsPayloadFunc returns true if any signal's payload satisfies eq.
func (g *Group) ContainsPayloadFunc(eq func(payload any) bool) bool {
	if g.HasChainableErr() || g.IsEmpty() {
		return false
	}
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
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
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
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
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
func (g *Group) With(signals ...*Signal) *Group {
	if g.HasChainableErr() {
		return g
	}
	newSignals := make(Signals, g.Len()+len(signals))
	copy(newSignals, g.signals)
	for i, sig := range signals {
		if sig == nil {
			return NewGroup().WithChainableErr(ErrInvalidSignal)
		}
		if sig.HasChainableErr() {
			return NewGroup().WithChainableErr(sig.ChainableErr())
		}
		newSignals[g.Len()+i] = sig
	}
	return newGroupFromSignals(newSignals)
}

// WithPayloads returns a new group with signals created from the given payloads appended.
func (g *Group) WithPayloads(payloads ...any) *Group {
	if g.HasChainableErr() {
		return g
	}
	newSignals := make(Signals, g.Len()+len(payloads))
	copy(newSignals, g.signals)
	for i, p := range payloads {
		newSignals[g.Len()+i] = New(p)
	}
	return newGroupFromSignals(newSignals)
}

// Join returns a new group containing signals from both groups.
func (g *Group) Join(other *Group) *Group {
	if g.HasChainableErr() {
		return g
	}
	if other.HasChainableErr() {
		return NewGroup().WithChainableErr(other.ChainableErr())
	}
	newSignals := make(Signals, g.Len()+other.Len())
	copy(newSignals, g.signals)
	copy(newSignals[g.Len():], other.signals)
	return newGroupFromSignals(newSignals)
}

// All returns a cloned slice of signals. The slice is independent of the group;
// the *Signal pointers inside are shared, but Signal is copy-on-write so callers
// cannot corrupt group state through the returned pointers.
func (g *Group) All() (Signals, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return slices.Clone(g.signals), nil
}

// WithChainableErr sets a chainable error and returns a new group.
func (g *Group) WithChainableErr(err error) *Group {
	return &Group{
		chainableErr: err,
		signals:      slices.Clone(g.signals),
	}
}

// HasChainableErr returns true when a chainable error is set.
func (g *Group) HasChainableErr() bool {
	return g.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (g *Group) ChainableErr() error {
	return g.chainableErr
}

// Len returns the number of signals in the group.
func (g *Group) Len() int {
	if g.HasChainableErr() {
		return 0
	}
	return len(g.signals)
}

// ForEach applies the action to each signal and returns a result group.
// On error the receiver is unchanged; the returned group carries the error.
func (g *Group) ForEach(action func(*Signal) error) *Group {
	if g.HasChainableErr() {
		return g
	}
	for _, s := range g.signals {
		if err := action(s); err != nil {
			return newGroupFromSignals(g.signals).WithChainableErr(err)
		}
	}
	return g
}

// ForEachIf applies the action only to signals that match the predicate.
func (g *Group) ForEachIf(predicate Predicate, action func(*Signal) error) *Group {
	if g.HasChainableErr() {
		return g
	}
	for _, s := range g.signals {
		if predicate(s) {
			if err := action(s); err != nil {
				return newGroupFromSignals(g.signals).WithChainableErr(err)
			}
		}
	}
	return g
}

// Filter returns a new group with signals that pass the predicate.
func (g *Group) Filter(p Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	filtered := make(Signals, 0, len(g.signals))
	for _, s := range g.signals {
		if p(s) {
			filtered = append(filtered, s)
		}
	}
	return newGroupFromSignals(filtered)
}

// Map returns a new group with every signal transformed by the mapper.
func (g *Group) Map(m Mapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := make(Signals, 0, len(g.signals))
	for _, s := range g.signals {
		mapped = append(mapped, m(cloneSignal(s)))
	}
	return newGroupFromSignals(mapped)
}

// MapIf is like Map but applies the mapper only to signals matching the predicate.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := make(Signals, len(g.signals))
	for i, s := range g.signals {
		clonedSignal := cloneSignal(s)
		if predicate(s) {
			mapped[i] = mapper(clonedSignal)
		} else {
			mapped[i] = clonedSignal
		}
	}
	return newGroupFromSignals(mapped)
}

// MapPayloads returns a new group with every payload transformed by the mapper.
func (g *Group) MapPayloads(mapper PayloadMapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := make(Signals, 0, len(g.signals))
	for _, s := range g.signals {
		mapped = append(mapped, s.MapPayload(mapper))
	}
	return newGroupFromSignals(mapped)
}

// MapPayloadsIf is like MapPayloads but applies the mapper only to signals matching the predicate.
func (g *Group) MapPayloadsIf(predicate Predicate, mapper PayloadMapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := make(Signals, len(g.signals))
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
	if g.HasChainableErr() {
		return initial
	}
	acc := initial
	for _, s := range g.signals {
		acc = fn(acc, s)
	}
	return acc
}

// ReducePayloads accumulates all signal payloads into a single value using the given function.
func (g *Group) ReducePayloads(initial any, fn PayloadReducer) any {
	if g.HasChainableErr() {
		return initial
	}
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
