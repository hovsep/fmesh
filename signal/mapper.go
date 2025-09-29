package signal

// Mapper transforms a Signal into a new Signal.
type Mapper func(signal *Signal) *Signal

// PayloadMapper transforms a payload into a new payload.
type PayloadMapper func(payload any) any
