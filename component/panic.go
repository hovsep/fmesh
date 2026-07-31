package component

import (
	"fmt"
)

// PanicError is what a recovered panic becomes.
//
// The stack is deliberately not part of Error(). Formatting a full debug.Stack()
// into the message produced multi-kilobyte single-line errors: unreadable in a
// log, useless in a test assertion, and impossible to grep. Error() is one line
// naming what was thrown; the stack is still there, reachable with errors.As.
//
//	var panicErr *component.PanicError
//	if errors.As(err, &panicErr) {
//	    log.Printf("%s in %s\n%s", panicErr, panicErr.ComponentName, panicErr.StackTrace())
//	}
type PanicError struct {
	// ComponentName is the component whose activation panicked. It is not part
	// of Error() because ActivationResult.ActivationErrorWithComponentName
	// already prefixes it; keeping it here makes it reachable programmatically
	// rather than only by parsing a message.
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
