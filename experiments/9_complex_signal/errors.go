package main

import (
	"errors"
	"sync"
)

type ErrorHandlingStrategy int

const (
	StopOnFirstError ErrorHandlingStrategy = iota
	IgnoreAll
)

var (
	errWaitingForInputResetInputs = errors.New("component is not ready (waiting for one or more inputs). All inputs will be reset")
	errWaitingForInputKeepInputs  = errors.New("component is not ready (waiting for one or more inputs). All inputs will be kept")
)

// ActivationResult defines the result (possibly an error) of the activation of given component
type ActivationResult struct {
	activated     bool
	componentName string
	err           error
}

// HopResult describes the outcome of every single component activation in single hop
type HopResult struct {
	sync.Mutex
	activationResults map[string]error
}

func (r *HopResult) hasErrors() bool {
	for _, err := range r.activationResults {
		if err != nil {
			return true
		}
	}
	return false
}

func isWaitingForInput(err error) bool {
	return errors.Is(err, errWaitingForInputResetInputs) || errors.Is(err, errWaitingForInputKeepInputs)
}
