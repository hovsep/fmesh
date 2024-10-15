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
		Chainable: &common.Chainable{},
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
