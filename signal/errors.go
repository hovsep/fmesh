package signal

import "errors"

var (
	ErrNoSignalsInGroup = errors.New("group has no signals")
	ErrInvalidSignal    = errors.New("signal is invalid")
)
