package signal

// Predicate is a function that tests whether a Signal matches a condition.
type Predicate func(signal *Signal) bool

// Mapper transforms a Signal into a new Signal.
type Mapper func(signal *Signal) *Signal

// PayloadMapper transforms a payload into a new payload.
type PayloadMapper func(payload any) any

// Signals is a slice of signals.
type Signals []*Signal
