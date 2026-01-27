package signal

import (
	"github.com/hovsep/fmesh/labels"
)

// Signal is a wrapper around the data flowing between components.
type Signal struct {
	chainableErr error
	labels       *labels.Collection
	payload      any
}

// New creates a new signal with the given payload.
// Signals carry data between components through ports.
// The payload can be any type (string, int, struct, slice, nil, etc.).
//
// Example (in activation function):
//
//	// Send a simple value
//	this.OutputByName("count").PutSignals(signal.New(42))
//
//	// Send a complex value
//	payload := map[string]interface{}{"status": "ok", "data": items}
//	this.OutputByName("result").PutSignals(signal.New(payload))
func New(payload any) *Signal {
	return &Signal{
		chainableErr: nil,
		labels:       labels.NewCollection(),
		payload:      payload,
	}
}

// Labels returns the signal's labels collection.
func (s *Signal) Labels() *labels.Collection {
	if s.HasChainableErr() {
		return labels.NewCollection().WithChainableErr(s.ChainableErr())
	}
	return s.labels
}

// SetLabels replaces all labels and returns the signal for chaining.
func (s *Signal) SetLabels(labelMap labels.Map) *Signal {
	if s.HasChainableErr() {
		return s
	}
	s.labels.Clear().AddMany(labelMap)
	return s
}

// AddLabels adds or updates labels and returns the signal for chaining.
func (s *Signal) AddLabels(labelMap labels.Map) *Signal {
	if s.HasChainableErr() {
		return s
	}
	s.labels.AddMany(labelMap)
	return s
}

// AddLabel adds or updates a single label and returns the signal for chaining.
func (s *Signal) AddLabel(name, value string) *Signal {
	if s.HasChainableErr() {
		return s
	}
	s.labels.Add(name, value)
	return s
}

// ClearLabels removes all labels and returns the signal for chaining.
func (s *Signal) ClearLabels() *Signal {
	if s.HasChainableErr() {
		return s
	}
	s.labels.Clear()
	return s
}

// WithoutLabels removes specific labels and returns the signal for chaining.
func (s *Signal) WithoutLabels(names ...string) *Signal {
	if s.HasChainableErr() {
		return s
	}
	s.labels.Without(names...)
	return s
}

// Payload returns the signal's payload.
func (s *Signal) Payload() (any, error) {
	if s.HasChainableErr() {
		return nil, s.ChainableErr()
	}
	return s.payload, nil
}

// PayloadOrNil returns payload or nil in case of error.
func (s *Signal) PayloadOrNil() any {
	return s.PayloadOrDefault(nil)
}

// PayloadOrDefault returns payload or provided default value in case of error.
func (s *Signal) PayloadOrDefault(defaultPayload any) any {
	payload, err := s.Payload()
	if err != nil {
		return defaultPayload
	}
	return payload
}

// WithChainableErr sets a chainable error and returns the signal.
func (s *Signal) WithChainableErr(err error) *Signal {
	s.chainableErr = err
	return s
}

// HasChainableErr returns true when a chainable error is set.
func (s *Signal) HasChainableErr() bool {
	return s.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (s *Signal) ChainableErr() error {
	return s.chainableErr
}

// Map applies a given mapper func and returns a new signal.
func (s *Signal) Map(m Mapper) *Signal {
	if s.HasChainableErr() {
		return s
	}
	return m(s)
}

// MapPayload applies a mapper function to the signal's payload and returns a new signal.
// The new signal preserves all labels from the original signal.
func (s *Signal) MapPayload(mapper PayloadMapper) *Signal {
	if s.HasChainableErr() {
		return s
	}

	payload, err := s.Payload()
	if err != nil {
		return New(nil).WithChainableErr(err)
	}

	// Create new signal with mapped payload
	newSignal := New(mapper(payload))

	// Copy labels using ForEach
	s.labels.ForEach(func(label, value string) error {
		return newSignal.AddLabel(label, value).ChainableErr()
	})

	return newSignal
}
