package component

import "errors"

// ActivationResult defines the result (possibly an error) of the activation of a given component in a given cycle.
type ActivationResult struct {
	componentName    string
	activated        bool
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

// NewActivationResult creates a new activation result for the given component.
func NewActivationResult(componentName string) *ActivationResult {
	return &ActivationResult{
		componentName: componentName,
	}
}

// ComponentName returns the name of the component this activation result belongs to.
func (ar *ActivationResult) ComponentName() string {
	return ar.componentName
}

// Activated returns true if the component was activated.
func (ar *ActivationResult) Activated() bool {
	return ar.activated
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

// SetActivated sets the activated flag and returns the activation result.
func (ar *ActivationResult) SetActivated(activated bool) *ActivationResult {
	ar.activated = activated
	return ar
}

// SetCode sets the activation code and returns the activation result.
func (ar *ActivationResult) SetCode(code ActivationResultCode) *ActivationResult {
	ar.code = code
	return ar
}

// AddError appends an error to the activation result and returns it.
func (ar *ActivationResult) AddError(err error) *ActivationResult {
	ar.activationErrors = append(ar.activationErrors, err)
	return ar
}

// newResult builds a finished result. Every code but NoInput means the component activated.
func (c *Component) newResult(code ActivationResultCode, errs ...error) *ActivationResult {
	return &ActivationResult{
		componentName:    c.name,
		activated:        code != ActivationCodeNoInput,
		code:             code,
		activationErrors: errs,
	}
}

func waitingCode(err error) ActivationResultCode {
	if errors.Is(err, ErrWaitKeepingInputs) {
		return ActivationCodeWaitingForInputsKeep
	}
	return ActivationCodeWaitingForInputsClear
}
