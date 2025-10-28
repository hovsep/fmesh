package signal

import (
	"github.com/hovsep/fmesh/labels"
)

// Group represents a list of signals.
type Group struct {
	chainableErr error
	signals      Signals
}

// NewGroup creates an empty group.
func NewGroup(payloads ...any) *Group {
	newGroup := &Group{
		chainableErr: nil,
	}

	signals := make(Signals, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroup.withSignals(signals)
}

// First returns the first signal in the group.
func (g *Group) First() *Signal {
	if g.HasChainableErr() {
		return New(nil).WithChainableErr(g.ChainableErr())
	}

	if g.IsEmpty() {
		g.WithChainableErr(ErrNoSignalsInGroup)
		return New(nil).WithChainableErr(g.ChainableErr())
	}

	return g.signals[0]
}

// IsEmpty returns true when there are no signals in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// AnyMatch returns true if at least one signal matches the predicate.
func (g *Group) AnyMatch(p Predicate) bool {
	if g.HasChainableErr() {
		return false
	}

	if g.IsEmpty() {
		return false
	}

	for _, sig := range g.signals {
		if p(sig) {
			return true
		}
	}

	return false
}

// AllMatch returns true if all signals match the predicate.
func (g *Group) AllMatch(p Predicate) bool {
	if g.HasChainableErr() {
		return false
	}

	if g.IsEmpty() {
		return false
	}

	for _, sig := range g.signals {
		if !p(sig) {
			return false
		}
	}

	return true
}

// FirstMatch returns the first signal that passes the predicate.
func (g *Group) FirstMatch(p Predicate) *Signal {
	if g.HasChainableErr() {
		return New(nil).WithChainableErr(g.ChainableErr())
	}

	if g.IsEmpty() {
		g.WithChainableErr(ErrNoSignalsInGroup)
		return New(nil).WithChainableErr(g.ChainableErr())
	}

	for _, sig := range g.signals {
		if p(sig) {
			return sig
		}
	}

	g.WithChainableErr(ErrNotFound)
	return New(nil).WithChainableErr(g.ChainableErr())
}

// FirstPayload returns the payload of the first signal.
func (g *Group) FirstPayload() (any, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}

	return g.First().Payload()
}

// FirstPayloadOrDefault returns the payload of the first signal or the provided default.
func (g *Group) FirstPayloadOrDefault(defaultPayload any) any {
	payload, err := g.FirstPayload()
	if err != nil {
		return defaultPayload
	}
	return payload
}

// FirstPayloadOrNil returns the payload of the first signal or nil in case of error.
func (g *Group) FirstPayloadOrNil() any {
	return g.FirstPayloadOrDefault(nil)
}

// AllPayloads returns a slice with all payloads of the all signals in the group.
func (g *Group) AllPayloads() ([]any, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}

	all := make([]any, g.Len())
	var err error
	for i, sig := range g.signals {
		all[i], err = sig.Payload()
		if err != nil {
			return nil, err
		}
	}
	return all, nil
}

// With returns the group with added signals.
func (g *Group) With(signals ...*Signal) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}

	newSignals := make(Signals, g.Len()+len(signals))
	copy(newSignals, g.signals)
	for i, sig := range signals {
		if sig == nil {
			g.WithChainableErr(ErrInvalidSignal)
			return NewGroup().WithChainableErr(g.ChainableErr())
		}

		if sig.HasChainableErr() {
			g.WithChainableErr(sig.ChainableErr())
			return NewGroup().WithChainableErr(g.ChainableErr())
		}

		newSignals[g.Len()+i] = sig
	}

	return g.withSignals(newSignals)
}

// Without removes signals matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}
	// Keep signals that DON'T match the predicate
	return g.Filter(func(s *Signal) bool {
		return !predicate(s)
	})
}

// WithPayloads returns a group with added signals created from provided payloads.
func (g *Group) WithPayloads(payloads ...any) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}

	newSignals := make(Signals, g.Len()+len(payloads))
	copy(newSignals, g.signals)
	for i, p := range payloads {
		newSignals[g.Len()+i] = New(p)
	}
	return g.withSignals(newSignals)
}

// withSignals sets signals.
func (g *Group) withSignals(signals Signals) *Group {
	g.signals = signals
	return g
}

// AllAsSlice returns signals as Signals wrapper type.
func (g *Group) AllAsSlice() (Signals, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.signals, nil
}

// AllAsSliceOrNil returns signals as Signals wrapper or nil in case of any error.
func (g *Group) AllAsSliceOrNil() Signals {
	return g.AllAsSliceOrDefault(nil)
}

// AllAsSliceOrDefault returns signals as Signals wrapper or default in case of any error.
func (g *Group) AllAsSliceOrDefault(defaultSignals Signals) Signals {
	signals, err := g.AllAsSlice()
	if err != nil {
		return defaultSignals
	}
	return signals
}

// WithChainableErr sets a chainable error and returns the group.
func (g *Group) WithChainableErr(err error) *Group {
	g.chainableErr = err
	return g
}

// HasChainableErr returns true when a chainable error is set.
func (g *Group) HasChainableErr() bool {
	return g.chainableErr != nil
}

// ChainableErr returns chainable error.
func (g *Group) ChainableErr() error {
	return g.chainableErr
}

// Len returns a number of signals in a group.
func (g *Group) Len() int {
	return len(g.signals)
}

// WithSignalLabels sets labels on each signal within the group and returns it.
func (g *Group) WithSignalLabels(labelMap labels.Map) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}

	for _, s := range g.AllAsSliceOrNil() {
		s.WithLabels(labelMap)
	}
	return g
}

// Filter returns a new group with signals that pass the filter.
func (g *Group) Filter(p Predicate) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}

	filteredSignals := make(Signals, 0)
	for _, s := range g.AllAsSliceOrNil() {
		if p(s) {
			filteredSignals = append(filteredSignals, s)
		}
	}

	return NewGroup().withSignals(filteredSignals)
}

// Map returns a new group with signals transformed by the mapper function.
func (g *Group) Map(m Mapper) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}

	mappedSignals := make(Signals, 0)
	for _, s := range g.AllAsSliceOrNil() {
		mappedSignals = append(mappedSignals, s.Map(m))
	}

	return NewGroup().withSignals(mappedSignals)
}

// MapPayloads returns a new group with payloads transformed by the mapper function.
func (g *Group) MapPayloads(mapper PayloadMapper) *Group {
	if g.HasChainableErr() {
		// Do nothing but propagate the error
		return g
	}

	mappedSignals := make(Signals, 0)
	for _, s := range g.AllAsSliceOrNil() {
		mappedSignals = append(mappedSignals, s.MapPayload(mapper))
	}

	return NewGroup().withSignals(mappedSignals)
}

// FirstOrDefault returns the first signal or the provided default.
func (g *Group) FirstOrDefault(defaultSignal *Signal) *Signal {
	if g.HasChainableErr() || g.IsEmpty() {
		return defaultSignal
	}
	return g.signals[0]
}

// FirstOrNil returns the first signal or nil.
func (g *Group) FirstOrNil() *Signal {
	if g.HasChainableErr() || g.IsEmpty() {
		return nil
	}
	return g.signals[0]
}

// NoneMatch returns true if no signals match the predicate.
func (g *Group) NoneMatch(predicate Predicate) bool {
	return !g.AnyMatch(predicate)
}

// CountMatch returns the number of signals that match the predicate.
func (g *Group) CountMatch(predicate Predicate) int {
	if g.HasChainableErr() {
		return 0
	}
	count := 0
	for _, sig := range g.signals {
		if predicate(sig) {
			count++
		}
	}
	return count
}
