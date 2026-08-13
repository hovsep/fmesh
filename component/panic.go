package component

import (
	"fmt"
)

// PanicError is what a recovered panic becomes.
//
// Error() is one line naming what was thrown; the stack is deliberately not in
// the message (a full stack made errors unreadable and ungreppable) and is
// reached with errors.As and StackTrace.
type PanicError struct {
	// ComponentName is the component whose activation panicked. Not part of
	// Error(), which the activation result already prefixes with it.
	ComponentName string

	// Value is whatever was passed to panic().
	Value any

	// Stack is the goroutine stack captured at the moment of recovery.
	Stack []byte
}

// Error implements error.
func (e *PanicError) Error() string {
	return fmt.Sprintf("panicked: %v", e.Value)
}

// StackTrace returns the goroutine stack captured when the panic was recovered.
func (e *PanicError) StackTrace() []byte {
	return e.Stack
}

// Unwrap exposes the panic value when something threw an error, so errors.Is and
// errors.As reach through the panic to whatever was actually thrown.
func (e *PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}
