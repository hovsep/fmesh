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

// Payloads returns all payloads
func (s *Signal) Payloads() []any {
	return s.payloads
}

// Payload returns the first payloads (useful when you are sure there is just one payloads)
// It panics when used with signal that carries multiple payloads
func (s *Signal) Payload() any {
	if s.Len() != 1 {
		panic("signal has zero or multiple payloads")
	}
	return s.payloads[0]
}

// Merge returns a new signal which payloads is combined from 2 original signals
func (s *Signal) Merge(anotherSignal *Signal) *Signal {
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
