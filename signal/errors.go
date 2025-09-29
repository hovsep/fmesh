package signal

import "errors"

var (
	// ErrNoSignalsInGroup is returned when a group has no signals
	ErrNoSignalsInGroup = errors.New("group has no signals")
	// ErrInvalidSignal is returned when a signal is invalid
	ErrInvalidSignal = errors.New("signal is invalid")
	// ErrNotFound is returned when no signal matching predicate is found
	ErrNotFound = errors.New("signal not found")
)
