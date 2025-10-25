package signal

import (
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/labels"
)

// Signals is a slice of signals.
type Signals []*Signal

// Group represents a list of signals.
type Group struct {
	*common.Chainable
	signals Signals
}

// NewGroup creates an empty group.
func NewGroup(payloads ...any) *Group {
	newGroup := &Group{
		Chainable: common.NewChainable(),
	}

	signals := make(Signals, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroup.withSignals(signals)
}

// FirstSignal returns the first signal in the group.
func (g *Group) FirstSignal() *Signal {
	if g.HasErr() {
		return New(nil).WithErr(g.Err())
	}

	if g.Len() == 0 {
		g.SetErr(ErrNoSignalsInGroup)
		return New(nil).WithErr(g.Err())
	}

	return g.signals[0]
}

// IsEmpty returns true when there are no signals in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// Any returns true if at least one signal matches the predicate.
func (g *Group) Any(p Predicate) bool {
	if g.HasErr() {
		return false
	}

	if g.Len() == 0 {
		return false
	}

	for _, sig := range g.signals {
		if p(sig) {
			return true
		}
	}

	return false
}

// All returns true if all signals match the predicate.
func (g *Group) All(p Predicate) bool {
	if g.HasErr() {
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

// First returns the first signal that passes the predicate.
func (g *Group) First(p Predicate) *Signal {
	if g.HasErr() {
		return New(nil).WithErr(g.Err())
	}

	if g.IsEmpty() {
		g.SetErr(ErrNoSignalsInGroup)
		return New(nil).WithErr(g.Err())
	}

	for _, sig := range g.signals {
		if p(sig) {
			return sig
		}
	}

	g.SetErr(ErrNotFound)
	return New(nil).WithErr(g.Err())
}

// FirstPayload returns the payload of the first signal.
func (g *Group) FirstPayload() (any, error) {
	if g.HasErr() {
		return nil, g.Err()
	}

	return g.FirstSignal().Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group.
func (g *Group) AllPayloads() ([]any, error) {
	if g.HasErr() {
		return nil, g.Err()
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
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	newSignals := make(Signals, g.Len()+len(signals))
	copy(newSignals, g.signals)
	for i, sig := range signals {
		if sig == nil {
			g.SetErr(ErrInvalidSignal)
			return NewGroup().WithErr(g.Err())
		}

		if sig.HasErr() {
			g.SetErr(sig.Err())
			return NewGroup().WithErr(g.Err())
		}

		newSignals[g.Len()+i] = sig
	}

	return g.withSignals(newSignals)
}

// WithPayloads returns a group with added signals created from provided payloads.
func (g *Group) WithPayloads(payloads ...any) *Group {
	if g.HasErr() {
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

// Signals getter.
func (g *Group) Signals() (Signals, error) {
	if g.HasErr() {
		return nil, g.Err()
	}
	return g.signals, nil
}

// SignalsOrNil returns signals or nil in case of any error.
func (g *Group) SignalsOrNil() Signals {
	return g.SignalsOrDefault(nil)
}

// SignalsOrDefault returns signals or default in case of any error.
func (g *Group) SignalsOrDefault(defaultSignals Signals) Signals {
	signals, err := g.Signals()
	if err != nil {
		return defaultSignals
	}
	return signals
}

// WithErr returns group with error.
func (g *Group) WithErr(err error) *Group {
	g.SetErr(err)
	return g
}

// Len returns a number of signals in a group.
func (g *Group) Len() int {
	return len(g.signals)
}

// WithSignalLabels sets labels on each signal within the group and returns it.
func (g *Group) WithSignalLabels(labels labels.Map) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	for _, s := range g.SignalsOrNil() {
		s.WithLabels(labels)
	}
	return g
}

// Filter returns a new group with signals that pass the filter.
func (g *Group) Filter(p Predicate) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	filteredSignals := make(Signals, 0)
	for _, s := range g.SignalsOrNil() {
		if p(s) {
			filteredSignals = append(filteredSignals, s)
		}
	}

	return NewGroup().withSignals(filteredSignals)
}

// Map returns a new group with signals transformed by the mapper function.
func (g *Group) Map(m Mapper) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	mappedSignals := make(Signals, 0)
	for _, s := range g.SignalsOrNil() {
		mappedSignals = append(mappedSignals, s.Map(m))
	}

	return NewGroup().withSignals(mappedSignals)
}

// MapPayloads returns a new group with payloads transformed by the mapper function.
func (g *Group) MapPayloads(mapper PayloadMapper) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	mappedSignals := make(Signals, 0)
	for _, s := range g.SignalsOrNil() {
		mappedSignals = append(mappedSignals, s.MapPayload(mapper))
	}

	return NewGroup().withSignals(mappedSignals)
}
