package signal

import (
	"errors"
	"github.com/hovsep/fmesh/common"
)

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
	if g.HasChainError() {
		return New(nil).WithChainError(g.ChainError())
	}

	if len(g.signals) == 0 {
		return New(nil).WithChainError(errors.New("group has no signals"))
	}

	return g.signals[0]
}

// FirstPayload returns the first signal payload
func (g *Group) FirstPayload() (any, error) {
	if g.HasChainError() {
		return nil, g.ChainError()
	}

	return g.First().Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group
func (g *Group) AllPayloads() ([]any, error) {
	if g.HasChainError() {
		return nil, g.ChainError()
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
	if g.HasChainError() {
		// Do nothing, but propagate error
		return g
	}

	newSignals := make(Signals, len(g.signals)+len(signals))
	copy(newSignals, g.signals)
	for i, sig := range signals {
		if sig == nil {
			return g.WithChainError(errors.New("signal is nil"))
		}

		if sig.HasChainError() {
			return g.WithChainError(sig.ChainError())
		}

		newSignals[len(g.signals)+i] = sig
	}

	return g.withSignals(newSignals)
}

// WithPayloads returns a group with added signals created from provided payloads
func (g *Group) WithPayloads(payloads ...any) *Group {
	if g.HasChainError() {
		// Do nothing, but propagate error
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
	if g.HasChainError() {
		return nil, g.ChainError()
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

// WithChainError returns group with error
func (g *Group) WithChainError(err error) *Group {
	g.SetChainError(err)
	return g
}

// Len returns number of signals in group
func (g *Group) Len() int {
	return len(g.signals)
}
