package signal

// Signal is a wrapper around the data flowing between components
type Signal struct {
	payload []any //Slice is used in order to support nil payload
}

// New creates a new signal from the given payloads
func New(payload any) *Signal {
	return &Signal{payload: []any{payload}}
}

// Payload getter
func (s *Signal) Payload() any {
	return s.payload[0]
}
