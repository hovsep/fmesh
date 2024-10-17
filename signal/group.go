package signal

import (
	"errors"
	"github.com/hovsep/fmesh/common"
)

// Group represents a list of signals
type Group struct {
	*common.Chainable
	signals []*Signal
}

// NewGroup creates empty group
func NewGroup(payloads ...any) *Group {
	newGroup := &Group{
		Chainable: common.NewChainable(),
	}

	signals := make([]*Signal, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return newGroup.withSignals(signals)
}

// First returns the first signal in the group
func (group *Group) First() *Signal {
	if group.HasChainError() {
		return New(nil).WithChainError(group.ChainError())
	}

	if len(group.signals) == 0 {
		return New(nil).WithChainError(errors.New("group has no signals"))
	}

	return group.signals[0]
}

// FirstPayload returns the first signal payload
func (group *Group) FirstPayload() (any, error) {
	if group.HasChainError() {
		return nil, group.ChainError()
	}

	return group.First().Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group
func (group *Group) AllPayloads() ([]any, error) {
	if group.HasChainError() {
		return nil, group.ChainError()
	}

	all := make([]any, len(group.signals))
	var err error
	for i, sig := range group.signals {
		all[i], err = sig.Payload()
		if err != nil {
			return nil, err
		}
	}
	return all, nil
}

// With returns the group with added signals
func (group *Group) With(signals ...*Signal) *Group {
	if group.HasChainError() {
		// Do nothing, but propagate error
		return group
	}

	newSignals := make([]*Signal, len(group.signals)+len(signals))
	copy(newSignals, group.signals)
	for i, sig := range signals {
		if sig == nil {
			return group.WithChainError(errors.New("signal is nil"))
		}

		if sig.HasChainError() {
			return group.WithChainError(sig.ChainError())
		}

		newSignals[len(group.signals)+i] = sig
	}

	return group.withSignals(newSignals)
}

// WithPayloads returns a group with added signals created from provided payloads
func (group *Group) WithPayloads(payloads ...any) *Group {
	if group.HasChainError() {
		// Do nothing, but propagate error
		return group
	}

	newSignals := make([]*Signal, len(group.signals)+len(payloads))
	copy(newSignals, group.signals)
	for i, p := range payloads {
		newSignals[len(group.signals)+i] = New(p)
	}
	return group.withSignals(newSignals)
}

// withSignals sets signals
func (group *Group) withSignals(signals []*Signal) *Group {
	group.signals = signals
	return group
}

// Signals getter
func (group *Group) Signals() ([]*Signal, error) {
	if group.HasChainError() {
		return nil, group.ChainError()
	}
	return group.signals, nil
}

// SignalsOrNil returns signals or nil in case of any error
func (group *Group) SignalsOrNil() []*Signal {
	return group.SignalsOrDefault(nil)
}

// SignalsOrDefault returns signals or default in case of any error
func (group *Group) SignalsOrDefault(defaultSignals []*Signal) []*Signal {
	signals, err := group.Signals()
	if err != nil {
		return defaultSignals
	}
	return signals
}

// WithChainError returns group with error
func (group *Group) WithChainError(err error) *Group {
	group.SetChainError(err)
	return group
}

// Len returns number of signals in group
func (group *Group) Len() int {
	return len(group.signals)
}
