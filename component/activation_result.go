package component

import "errors"

// ActivationResult is how one component's activation in one cycle ended.
type ActivationResult struct {
	componentName    string
	code             ActivationResultCode
	activationErrors []error // All errors accumulated during activation (component error + any hook errors)
}

// ActivationResultCode denotes specific info about how a component been activated or why not activated at all.
type ActivationResultCode int

func (a ActivationResultCode) String() string {
	switch a {
	case ActivationCodeUndefined:
		return "Undefined"
	case ActivationCodeOK:
		return "Success"
	case ActivationCodeNoInput:
		return "No input"
	case ActivationCodeReturnedError:
		return "Finished with error"
	case ActivationCodePanicked:
		return "Finished with panic"
	case ActivationCodeWaitingForInputsClear:
		return "Waiting for input (clear)"
	case ActivationCodeWaitingForInputsKeep:
		return "Waiting for input (keep)"
	case ActivationCodeHookFailed:
		return "Hook failed"
	default:
		return "Unknown code"
	}
}

const (
	// ActivationCodeUndefined - the zero value: no code has been set.
	ActivationCodeUndefined ActivationResultCode = iota

	// ActivationCodeOK - component is activated and did not return any errors.
	ActivationCodeOK

	// ActivationCodeNoInput - component is not activated because it has no input set.
	ActivationCodeNoInput

	// ActivationCodeReturnedError - component is activated but returned an error.
	ActivationCodeReturnedError

	// ActivationCodePanicked - component is activated, but panicked.
	ActivationCodePanicked

	// ActivationCodeWaitingForInputsClear - the component waits for specific inputs, but all input signals in the current activation cycle may be cleared (default behavior).
	ActivationCodeWaitingForInputsClear

	// ActivationCodeWaitingForInputsKeep - the component waits for signals on specific input ports and wants to keep current input signals for the next cycle.
	ActivationCodeWaitingForInputsKeep

	// ActivationCodeHookFailed - a hook failed, preventing or disrupting activation.
	ActivationCodeHookFailed
)

// NewActivationResult creates the result of one activation of the named component.
func NewActivationResult(componentName string, code ActivationResultCode, errs ...error) *ActivationResult {
	return &ActivationResult{
		componentName:    componentName,
		code:             code,
		activationErrors: errs,
	}
}

// ComponentName returns the name of the component this activation result belongs to.
func (ar *ActivationResult) ComponentName() string {
	return ar.componentName
}

// Activated reports whether the activation ran: every code but NoInput and Undefined.
func (ar *ActivationResult) Activated() bool {
	return ar.code != ActivationCodeNoInput && ar.code != ActivationCodeUndefined
}

// Err returns all accumulated activation errors joined into one, or nil if there are none.
func (ar *ActivationResult) Err() error {
	return errors.Join(ar.activationErrors...)
}

// Errors returns all accumulated activation errors.
func (ar *ActivationResult) Errors() []error {
	return ar.activationErrors
}

// Code returns the activation result code.
func (ar *ActivationResult) Code() ActivationResultCode {
	return ar.code
}

// IsError returns true when an activation result has an error.
// Hook failures count as errors so they surface through the error handling strategy.
func (ar *ActivationResult) IsError() bool {
	return (ar.code == ActivationCodeReturnedError || ar.code == ActivationCodeHookFailed) && len(ar.activationErrors) > 0
}

// IsPanic returns true when an activation result is derived from panic.
func (ar *ActivationResult) IsPanic() bool {
	return ar.code == ActivationCodePanicked && len(ar.activationErrors) > 0
}

// IsWaiting reports whether the component was waiting for inputs, in either mode.
func (ar *ActivationResult) IsWaiting() bool {
	return ar.code == ActivationCodeWaitingForInputsClear || ar.code == ActivationCodeWaitingForInputsKeep
}

// KeepsInputs reports whether the component is waiting and keeping its inputs for the next cycle.
func (ar *ActivationResult) KeepsInputs() bool {
	return ar.code == ActivationCodeWaitingForInputsKeep
}

func waitingCode(err error) ActivationResultCode {
	if errors.Is(err, ErrWaitKeepingInputs) {
		return ActivationCodeWaitingForInputsKeep
	}
	return ActivationCodeWaitingForInputsClear
}
