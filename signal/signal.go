package signal

// Signal describes a piece of data sent between components
type Signal struct {
	payload []any //Signal can carry multiple payloads (e.g. when multiple signals are combined)
}

// New creates a new signal from the given payloads
func New(payload ...any) *Signal {
	return &Signal{payload: payload}
}

// Len returns a number of payloads
func (s *Signal) Len() int {
	return len(s.payload)
}

// Payload returns all payloads
func (s *Signal) Payload() []any {
	return s.payload
}

// Merge returns a new signal which payload is combined from 2 original signals
func (s *Signal) Merge(anotherSignal *Signal) *Signal {
	//Merging with nothing
	if anotherSignal == nil || anotherSignal.Payload() == nil {
		return s
	}

	//Original signal is empty
	if s.Payload() == nil {
		return anotherSignal
	}

	return New(append(s.Payload(), anotherSignal.Payload()...)...)
}
