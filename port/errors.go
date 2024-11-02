package port

import (
	"errors"
)

var (
	ErrPortNotFoundInCollection    = errors.New("port not found")
	ErrInvalidRangeForIndexedGroup = errors.New("start index can not be greater than end index")
	ErrNilPort                     = errors.New("port is nil")
	ErrMissingLabel                = errors.New("port is missing required label")
	ErrInvalidPipeDirection        = errors.New("pipe must go from output to input")
)
