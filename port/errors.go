package port

import "errors"

var (
	ErrPortNotFoundInCollection    = errors.New("port not found")
	ErrPortNotReadyForFlush        = errors.New("port is not ready for flush")
	ErrInvalidRangeForIndexedGroup = errors.New("start index can not be greater than end index")
)
