package signal

// Mapper is a function that can be used to transform signals
type Mapper func(signal *Signal) *Signal
