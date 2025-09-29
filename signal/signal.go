package signal

import "github.com/hovsep/fmesh/common"

// Signal is a wrapper around the data flowing between components
type Signal struct {
	*common.Chainable
	common.LabeledEntity
	payload []any // Slice is used in order to support nil payload
}

// New creates a new signal from the given payloads
func New(payload any) *Signal {
	return &Signal{
		Chainable:     common.NewChainable(),
		LabeledEntity: common.NewLabeledEntity(nil),
		payload:       []any{payload},
	}
}

// Payload getter
func (s *Signal) Payload() (any, error) {
	if s.HasErr() {
		return nil, s.Err()
	}
	return s.payload[0], nil
}

// PayloadOrNil returns payload or nil in case of error
func (s *Signal) PayloadOrNil() any {
	return s.PayloadOrDefault(nil)
}

// PayloadOrDefault returns payload or provided default value in case of error
func (s *Signal) PayloadOrDefault(defaultPayload any) any {
	payload, err := s.Payload()
	if err != nil {
		return defaultPayload
	}
	return payload
}

// WithErr returns signal with error
func (s *Signal) WithErr(err error) *Signal {
	s.SetErr(err)
	return s
}

// WithLabels sets labels and returns the signal
func (s *Signal) WithLabels(labels common.LabelsCollection) *Signal {
	if s.HasErr() {
		return s
	}

	s.SetLabels(labels)
	return s
}

// Map applies a given mapper func and returns a new signal
func (s *Signal) Map(mapper Mapper) *Signal {
	if s.HasErr() {
		return s
	}
	return mapper(s)
}

// MapPayload sets labels and returns the signal
func (s *Signal) MapPayload(mapper PayloadMapper) *Signal {
	if s.HasErr() {
		return s
	}
	payload, err := s.Payload()
	if err != nil {
		return New(nil).WithErr(err)
	}
	return New(mapper(payload))
}
