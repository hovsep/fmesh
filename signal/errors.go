package signal

import "errors"

var (
	// ErrNoSignalsInGroup is returned when a group has no signals.
	ErrNoSignalsInGroup = errors.New("group has no signals")
	// ErrPayloadNotComparable is returned when a payload type is not comparable.
	ErrPayloadNotComparable = errors.New("payload type is not comparable, use ContainsPayloadFunc instead")
	// ErrScalarNotFoundInGroup is returned when no signal in the group has the requested scalar.
	ErrScalarNotFoundInGroup = errors.New("no signal in the group has the scalar")
)
