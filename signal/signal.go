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
