package signal

import (
	"maps"
	"slices"

	"github.com/hovsep/fmesh/labels"
)

// Signal is a wrapper around the data flowing between components.
// Mutating-style methods return a new *Signal; receivers are never modified.
type Signal struct {
	chainableErr error
	labels       *labels.Collection
	payload      []any // Slice is used in order to support nil payload
}

// cloneLabels returns an independent labels.Collection with the same entries.
func cloneLabels(c *labels.Collection) *labels.Collection {
	if c == nil {
		return labels.NewCollection()
	}
	if c.HasChainableErr() {
		return labels.NewCollection().WithChainableErr(c.ChainableErr())
	}
	return c.Filter(func(string, string) bool { return true })
}

// cloneSignal returns a deep copy suitable for aliasing-free group operations.
func cloneSignal(s *Signal) *Signal {
	if s == nil {
		return nil
	}
	return &Signal{
		chainableErr: s.chainableErr,
		labels:       cloneLabels(s.labels),
		payload:      slices.Clone(s.payload),
	}
}

// New creates a new signal with the given payload.
func New(payload any) *Signal {
	return &Signal{
		chainableErr: nil,
		labels:       labels.NewCollection(),
		payload:      []any{payload},
	}
}

// Labels returns a defensive copy of the signal's labels collection.
func (s *Signal) Labels() *labels.Collection {
	if s.HasChainableErr() {
		return labels.NewCollection().WithChainableErr(s.ChainableErr())
	}
	return cloneLabels(s.labels)
}

// SetLabels replaces all labels and returns a new signal.
func (s *Signal) SetLabels(labelMap labels.Map) *Signal {
	if s.HasChainableErr() {
		return s
	}
	next := s.cloneForMutation()
	next.labels = labels.NewCollection().AddMany(maps.Clone(labelMap))
	return next
}

// AddLabels adds or updates labels and returns a new signal.
func (s *Signal) AddLabels(labelMap labels.Map) *Signal {
	if s.HasChainableErr() {
		return s
	}
	next := s.cloneForMutation()
	next.labels = next.labels.AddMany(maps.Clone(labelMap))
	return next
}

// AddLabel adds or updates a single label and returns a new signal.
func (s *Signal) AddLabel(name, value string) *Signal {
	if s.HasChainableErr() {
		return s
	}
	next := s.cloneForMutation()
	next.labels = next.labels.Add(name, value)
	return next
}

// ClearLabels removes all labels and returns a new signal.
func (s *Signal) ClearLabels() *Signal {
	if s.HasChainableErr() {
		return s
	}
	next := s.cloneForMutation()
	next.labels = labels.NewCollection()
	return next
}

// WithoutLabels removes specific labels and returns a new signal.
func (s *Signal) WithoutLabels(names ...string) *Signal {
	if s.HasChainableErr() {
		return s
	}
	next := s.cloneForMutation()
	next.labels = next.labels.Without(names...)
	return next
}

func (s *Signal) cloneForMutation() *Signal {
	return &Signal{
		chainableErr: s.chainableErr,
		labels:       cloneLabels(s.labels),
		payload:      slices.Clone(s.payload),
	}
}

// Payload returns the signal's payload.
func (s *Signal) Payload() (any, error) {
	if s.HasChainableErr() {
		return nil, s.ChainableErr()
	}
	return s.payload[0], nil
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

// WithChainableErr sets a chainable error and returns a new signal.
func (s *Signal) WithChainableErr(err error) *Signal {
	return &Signal{
		chainableErr: err,
		labels:       cloneLabels(s.labels),
		payload:      slices.Clone(s.payload),
	}
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

	out := New(mapper(payload))
	out.labels = cloneLabels(s.labels)
	return out
}
