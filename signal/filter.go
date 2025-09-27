package signal

// Filter is a predicate used to filter signals. true means keep signal, false - drop
type Filter func(signal *Signal) bool
