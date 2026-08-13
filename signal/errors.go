package signal

import "errors"

var (
	// ErrNoSignalsInGroup is returned when a group has no signals.
	ErrNoSignalsInGroup = errors.New("group has no signals")
	// ErrPayloadNotComparable is returned when a payload type is not comparable.
	ErrPayloadNotComparable = errors.New("payload type is not comparable, use ContainsPayloadFunc instead")
)
