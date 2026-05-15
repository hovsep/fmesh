package signal

import "slices"

// Group represents a list of signals.
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

// NewGroup creates an empty group.
func NewGroup(payloads ...any) *Group {
	signals := make(Signals, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroupFromSignals(signals)
}

// First returns the first signal in the group.
func (g *Group) First() *Signal {
	if g.HasChainableErr() {
		return nil
	}
	if g.IsEmpty() {
		return nil
	}
	return g.signals[0]
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

// AnyMatch returns true if at least one signal matches the predicate.
func (g *Group) AnyMatch(p Predicate) bool {
	if g.HasChainableErr() || g.IsEmpty() {
		return false
	}
	return slices.ContainsFunc(g.signals, p)
}

// AllMatch returns true if all signals match the predicate.
func (g *Group) AllMatch(p Predicate) bool {
	if g.HasChainableErr() || g.IsEmpty() {
		return false
	}
	for _, sig := range g.signals {
		if !p(sig) {
			return false
		}
	}
	return true
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

// Add returns a new group with added signals. The receiver is never modified.
func (g *Group) Add(signals ...*Signal) *Group {
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

// Without removes signals matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	return g.Filter(func(s *Signal) bool {
		return !predicate(s)
	})
}

// AddFromPayloads returns a new group with added signals created from provided payloads.
func (g *Group) AddFromPayloads(payloads ...any) *Group {
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

// All returns a copy of all signals in the group.
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

// Filter returns a new group with signals that pass the filter.
func (g *Group) Filter(p Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	filtered := make(Signals, 0)
	for _, s := range g.signals {
		if p(s) {
			filtered = append(filtered, s)
		}
	}
	return newGroupFromSignals(filtered)
}

// Map returns a new group with signals transformed by the mapper function.
func (g *Group) Map(m Mapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := make(Signals, 0, len(g.signals))
	for _, s := range g.signals {
		mapped = append(mapped, s.Map(m))
	}
	return newGroupFromSignals(mapped)
}

// MapIf is like Map but only applies the mapper function to signals that match the predicate.
func (g *Group) MapIf(predicate Predicate, mapper Mapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := make(Signals, len(g.signals))
	for i, s := range g.signals {
		if predicate(s) {
			mapped[i] = s.Map(mapper)
		} else {
			mapped[i] = cloneSignal(s)
		}
	}
	return newGroupFromSignals(mapped)
}

// MapPayloadsIf is like MapPayloads but only applies the mapper to signals that match the predicate.
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

// MapPayloads returns a new group with payloads transformed by the mapper function.
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

// CountMatch returns the number of signals that match the predicate.
func (g *Group) CountMatch(predicate Predicate) int {
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
