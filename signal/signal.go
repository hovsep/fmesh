package signal

import "github.com/hovsep/fmesh/common"

// Signal is a wrapper around the data flowing between components
type Signal struct {
	*common.Chainable
	payload []any //Slice is used in order to support nil payload
}

// New creates a new signal from the given payloads
func New(payload any) *Signal {
	return &Signal{
		Chainable: common.NewChainable(),
		payload:   []any{payload},
	}
}

// Payload getter
func (s *Signal) Payload() (any, error) {
	if s.HasError() {
		return nil, s.Error()
	}
	return s.payload[0], nil
}

// PayloadOrNil returns payload or nil in case of error
func (s *Signal) PayloadOrNil() any {
	return s.PayloadOrDefault(nil)
}

// PayloadOrDefault returns payload or provided default value in case of error
func (s *Signal) PayloadOrDefault(defaultValue any) any {
	payload, err := s.Payload()
	if err != nil {
		return defaultValue
	}
	return payload
}

// WithError returns signal with error
func (s *Signal) WithError(err error) *Signal {
	s.SetError(err)
	return s
}
