package signal

import (
	"github.com/hovsep/fmesh/common"
)

// Signals is a slice of signals
type Signals []*Signal

// Group represents a list of signals
type Group struct {
	*common.Chainable
	signals Signals
}

// NewGroup creates empty group
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

// First returns the first signal in the group
func (g *Group) First() *Signal {
	if g.HasErr() {
		return New(nil).WithErr(g.Err())
	}

	if len(g.signals) == 0 {
		g.SetErr(ErrNoSignalsInGroup)
		return New(nil).WithErr(g.Err())
	}

	return g.signals[0]
}

// FirstPayload returns the first signal payload
func (g *Group) FirstPayload() (any, error) {
	if g.HasErr() {
		return nil, g.Err()
	}

	return g.First().Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group
func (g *Group) AllPayloads() ([]any, error) {
	if g.HasErr() {
		return nil, g.Err()
	}

	all := make([]any, len(g.signals))
	var err error
	for i, sig := range g.signals {
		all[i], err = sig.Payload()
		if err != nil {
			return nil, err
		}
	}
	return all, nil
}

// With returns the group with added signals
func (g *Group) With(signals ...*Signal) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	newSignals := make(Signals, len(g.signals)+len(signals))
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

		newSignals[len(g.signals)+i] = sig
	}

	return g.withSignals(newSignals)
}

// WithPayloads returns a group with added signals created from provided payloads
func (g *Group) WithPayloads(payloads ...any) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	newSignals := make(Signals, len(g.signals)+len(payloads))
	copy(newSignals, g.signals)
	for i, p := range payloads {
		newSignals[len(g.signals)+i] = New(p)
	}
	return g.withSignals(newSignals)
}

// withSignals sets signals
func (g *Group) withSignals(signals Signals) *Group {
	g.signals = signals
	return g
}

// Signals getter
func (g *Group) Signals() (Signals, error) {
	if g.HasErr() {
		return nil, g.Err()
	}
	return g.signals, nil
}

// SignalsOrNil returns signals or nil in case of any error
func (g *Group) SignalsOrNil() Signals {
	return g.SignalsOrDefault(nil)
}

// SignalsOrDefault returns signals or default in case of any error
func (g *Group) SignalsOrDefault(defaultSignals Signals) Signals {
	signals, err := g.Signals()
	if err != nil {
		return defaultSignals
	}
	return signals
}

// WithErr returns group with error
func (g *Group) WithErr(err error) *Group {
	g.SetErr(err)
	return g
}

// Len returns number of signals in group
func (g *Group) Len() int {
	return len(g.signals)
}

// WithSignalLabels sets labels on each signal within the group and returns it
func (g *Group) WithSignalLabels(labels common.LabelsCollection) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	for _, s := range g.SignalsOrNil() {
		s.WithLabels(labels)
	}
	return g
}

// Filter returns a new group with signals that pass the filter
func (g *Group) Filter(filter Filter) *Group {
	if g.HasErr() {
		// Do nothing but propagate the error
		return g
	}

	filteredSignals := make(Signals, 0)
	for _, s := range g.SignalsOrNil() {
		if filter(s) {
			filteredSignals = append(filteredSignals, s)
		}
	}

	return NewGroup().withSignals(filteredSignals)
}
