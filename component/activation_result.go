package component

import (
	"errors"
	"fmt"
)

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
	// ActivationCodeUndefined - used for error handling as zero instance.
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

// ActivationError returns all accumulated activation errors joined into a single error, or nil if there are none.
func (ar *ActivationResult) ActivationError() error {
	return errors.Join(ar.activationErrors...)
}

// ActivationErrors returns all accumulated activation errors as a slice.
func (ar *ActivationResult) ActivationErrors() []error {
	return ar.activationErrors
}

// ActivationErrorWithComponentName returns activation error enriched with component name.
func (ar *ActivationResult) ActivationErrorWithComponentName() error {
	return fmt.Errorf("component %s has activation error: %w", ar.componentName, ar.ActivationError())
}

// Code returns the activation result code.
func (ar *ActivationResult) Code() ActivationResultCode {
	return ar.code
}

// IsError returns true when an activation result has an error.
func (ar *ActivationResult) IsError() bool {
	return ar.code == ActivationCodeReturnedError && len(ar.activationErrors) > 0
}

// IsPanic returns true when an activation result is derived from panic.
func (ar *ActivationResult) IsPanic() bool {
	return ar.code == ActivationCodePanicked && len(ar.activationErrors) > 0
}

// SetActivated sets the activated flag and returns the activation result.
func (ar *ActivationResult) SetActivated(activated bool) *ActivationResult {
	ar.activated = activated
	return ar
}

// SetActivationCode sets the activation code and returns the activation result.
func (ar *ActivationResult) SetActivationCode(code ActivationResultCode) *ActivationResult {
	ar.code = code
	return ar
}

// WithActivationError appends an error to the activation result's error list and returns the activation result.
func (ar *ActivationResult) WithActivationError(activationError error) *ActivationResult {
	ar.activationErrors = append(ar.activationErrors, activationError)
	return ar
}

func (c *Component) newActivationResultOK() *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		SetActivationCode(ActivationCodeOK)
}

func (c *Component) newActivationResultNoInput() *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(false).
		SetActivationCode(ActivationCodeNoInput)
}

func (c *Component) newActivationResultReturnedError(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		SetActivationCode(ActivationCodeReturnedError).
		WithActivationError(fmt.Errorf("component returned an error: %w", err))
}

func (c *Component) newActivationResultPanicked(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		SetActivationCode(ActivationCodePanicked).
		WithActivationError(err)
}

// newActivationResultHookFailed builds a specific activation result for when a hook fails.
func (c *Component) newActivationResultHookFailed(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(false).
		SetActivationCode(ActivationCodeHookFailed).
		WithActivationError(err)
}

func (c *Component) newActivationResultWaitingForInputs(err error) *ActivationResult {
	activationCode := ActivationCodeWaitingForInputsClear
	if errors.Is(err, ErrWaitingForInputsKeep) {
		activationCode = ActivationCodeWaitingForInputsKeep
	}
	return NewActivationResult(c.Name()).
		SetActivated(true).
		SetActivationCode(activationCode).
		WithActivationError(err)
}

// IsWaitingForInput returns true if the component was waiting for inputs.
func IsWaitingForInput(activationResult *ActivationResult) bool {
	return activationResult.Code() == ActivationCodeWaitingForInputsClear ||
		activationResult.Code() == ActivationCodeWaitingForInputsKeep
}

// WantsToKeepInputs returns true if the component wants to keep inputs.
func WantsToKeepInputs(activationResult *ActivationResult) bool {
	return activationResult.Code() == ActivationCodeWaitingForInputsKeep
}
