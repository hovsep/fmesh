package signal

// Signal describes a piece of data sent between components
type Signal struct {
	payloads []any //Signal can carry multiple payloads (e.g. when multiple signals are combined)
}

// New creates a new signal from the given payloads
func New(payloads ...any) *Signal {
	return &Signal{payloads: payloads}
}

// Len returns a number of payloads
func (s *Signal) Len() int {
	return len(s.payloads)
}

// HasPayload must be used to check whether signal carries at least 1 payload
func (s *Signal) HasPayload() bool {
	return s.Len() > 0
}

// Payloads returns all payloads
func (s *Signal) Payloads() []any {
	return s.payloads
}

// Payload returns the first payload (useful when you are sure signal has only one payload)
// It panics when used with signal that carries multiple payloads
func (s *Signal) Payload() any {
	if s.Len() != 1 {
		panic("signal has zero or multiple payloads")
	}
	return s.payloads[0]
}

// Combine returns a new signal with combined payloads of 2 original signals
func (s *Signal) Combine(anotherSignal *Signal) *Signal {
	//Merging with nothing
	if anotherSignal == nil || anotherSignal.Payloads() == nil {
		return s
	}

	//Original signal is empty
	if s.Payloads() == nil {
		return anotherSignal
	}

	return New(append(s.Payloads(), anotherSignal.Payloads()...)...)
}
