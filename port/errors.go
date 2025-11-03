package port

import (
	"errors"
)

var (
	// ErrPortNotFoundInCollection is returned when a port is not found in a collection.
	ErrPortNotFoundInCollection = errors.New("port not found")
	// ErrInvalidRangeForIndexedGroup is returned when the start index is greater than the end index.
	ErrInvalidRangeForIndexedGroup = errors.New("start index can not be greater than end index")
	// ErrNilPort is returned when a port is nil.
	ErrNilPort = errors.New("port is nil")
	// ErrInvalidPipeDirection is returned when a pipe has an invalid direction.
	ErrInvalidPipeDirection = errors.New("pipe must go from output to input")
	// ErrWrongPortDirection is returned when a port has the wrong direction for the operation.
	ErrWrongPortDirection = errors.New("port has wrong direction")
	// ErrNoPortsInGroup is returned when a group has no ports.
	ErrNoPortsInGroup = errors.New("no ports in group")
	// ErrNoPortsInCollection is returned when a collection has no ports.
	ErrNoPortsInCollection = errors.New("no ports in collection")
	// ErrNoPortMatchesPredicate is returned when no port matches the predicate.
	ErrNoPortMatchesPredicate = errors.New("no port matches the predicate")
	// errUnexpectedErrorGettingPort is an internal error for unexpected situations.
	errUnexpectedErrorGettingPort = errors.New("unexpected error getting port")
)
