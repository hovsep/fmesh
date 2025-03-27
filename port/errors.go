package port

import (
	"errors"
)

var (
	// ErrPortNotFoundInCollection is returned when a port is not found in a collection
	ErrPortNotFoundInCollection    = errors.New("port not found")
	// ErrInvalidRangeForIndexedGroup is returned when the start index is greater than the end index
	ErrInvalidRangeForIndexedGroup = errors.New("start index can not be greater than end index")
	// ErrNilPort is returned when a port is nil
	ErrNilPort                     = errors.New("port is nil")
	// ErrMissingLabel is returned when a port is missing a required label
	ErrMissingLabel                = errors.New("port is missing required label")
	// ErrInvalidPipeDirection is returned when a pipe has an invalid direction
	ErrInvalidPipeDirection        = errors.New("pipe must go from output to input")
)
