package port

import (
	"errors"
)

var (
	// ErrInvalidRangeForIndexedGroup is returned when the start index is greater than the end index.
	ErrInvalidRangeForIndexedGroup = errors.New("start index can not be greater than end index")
	// ErrNilPort is returned when a port is nil.
	ErrNilPort = errors.New("port is nil")
	// ErrInvalidPipeDirection is returned when a pipe has an invalid direction.
	ErrInvalidPipeDirection = errors.New("pipe must go from output to input")
	// ErrWrongPortDirection is returned when a port has the wrong direction for the operation.
	ErrWrongPortDirection = errors.New("port has wrong direction")
)
