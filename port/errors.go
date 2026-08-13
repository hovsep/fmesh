package port

import (
	"errors"
)

var (
	// ErrInvalidRangeForIndexedGroup is returned when the start index is greater than the end index.
	ErrInvalidRangeForIndexedGroup = errors.New("start index can not be greater than end index")
	// ErrNilPort reports a nil port, either passed to pipe validation or used as
	// the receiver of a port method. Both have the same usual cause, so both name
	// it: a lookup by a name no port has.
	ErrNilPort = errors.New("port is nil: InputByName, OutputByName and Collection.ByName return nil for a name no port has")
	// ErrInvalidPipeDirection is returned when a pipe has an invalid direction.
	ErrInvalidPipeDirection = errors.New("pipe must go from output to input")
	// ErrWrongPortDirection is returned when a port has the wrong direction for the operation.
	ErrWrongPortDirection = errors.New("port has wrong direction")
)
