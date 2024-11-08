package cycle

import "errors"

var (
	errNoCyclesInGroup = errors.New("group has no cycles")
)
