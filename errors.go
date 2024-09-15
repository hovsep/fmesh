package fmesh

type ErrorHandlingStrategy int

const (
	StopOnFirstError ErrorHandlingStrategy = iota
	StopOnFirstPanic
	IgnoreAll
)
