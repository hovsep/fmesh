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
	signals := make([]*Signal, len(payloads))
	for i, payload := range payloads {
		signals[i] = New(payload)
	}
	return &Group{
		Chainable: &common.Chainable{},
		signals:   signals,
	}
}

// First returns the first signal in the group
func (group *Group) First() *Signal {
	if group.HasError() {
		sig := New(nil)
		sig.SetError(group.Error())
		return sig
	}

	return group.signals[0]
}

// FirstPayload returns the first signal payload
func (group *Group) FirstPayload() (any, error) {
	if group.HasError() {
		return nil, group.Error()
	}

	return group.First().Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group
func (group *Group) AllPayloads() ([]any, error) {
	if group.HasError() {
		return nil, group.Error()
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
	if group.HasError() {
		// Do nothing, but propagate error
		return group
	}

	newSignals := make([]*Signal, len(group.signals)+len(signals))
	copy(newSignals, group.signals)
	for i, sig := range signals {
		if sig == nil {
			group.SetError(errors.New("signal is nil"))
			return group
		}

		if sig.HasError() {
			group.SetError(sig.Error())
			return group
		}

		newSignals[len(group.signals)+i] = sig
	}

	return group.withSignals(newSignals)
}

// WithPayloads returns a group with added signals created from provided payloads
func (group *Group) WithPayloads(payloads ...any) *Group {
	if group.HasError() {
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
	if group.HasError() {
		return nil, group.Error()
	}
	return group.signals, nil
}

// SignalsOrNil returns signals or nil in case of any error
func (group *Group) SignalsOrNil() []*Signal {
	if group.HasError() {
		return nil
	}
	return group.signals
}

// WithError returns group with error
func (group *Group) WithError(err error) *Group {
	group.SetError(err)
	return group
}
