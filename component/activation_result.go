package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
)

// ActivationResult defines the result (possibly an error) of the activation of given component in given cycle
type ActivationResult struct {
	*common.Chainable
	componentName   string
	activated       bool
	code            ActivationResultCode
	activationError error // Error returned from component activation function
}

// ActivationResultCode denotes a specific info about how a component been activated or why not activated at all
type ActivationResultCode int

func (a ActivationResultCode) String() string {
	switch a {
	case ActivationCodeUndefined:
		return "UNDEFINED"
	case ActivationCodeOK:
		return "OK"
	case ActivationCodeNoInput:
		return "No input"
	case ActivationCodeNoFunction:
		return "Activation function is missing"
	case ActivationCodeReturnedError:
		return "Returned error"
	case ActivationCodePanicked:
		return "Panicked"
	case ActivationCodeWaitingForInputsClear:
		return "Component is waiting for input"
	case ActivationCodeWaitingForInputsKeep:
		return "Component is waiting for input and wants to keep all inputs till next cycle"
	default:
		return "Unsupported code"
	}
}

const (
	// ActivationCodeUndefined : used for error handling as zero instance
	ActivationCodeUndefined ActivationResultCode = iota

	// ActivationCodeOK : component is activated and did not return any errors
	ActivationCodeOK

	// ActivationCodeNoInput : component is not activated because it has no input set
	ActivationCodeNoInput

	// ActivationCodeNoFunction : component activation function is not set, so we can not activate it
	ActivationCodeNoFunction

	// ActivationCodeReturnedError : component is activated, but returned an error
	ActivationCodeReturnedError

	// ActivationCodePanicked : component is activated, but panicked
	ActivationCodePanicked

	// ActivationCodeWaitingForInputsClear : component waits for specific inputs, but all input signals in current activation cycle may be cleared (default behavior)
	ActivationCodeWaitingForInputsClear

	// ActivationCodeWaitingForInputsKeep : component waits for specific inputs, but wants to keep current input signals for the next cycle
	ActivationCodeWaitingForInputsKeep
)

// NewActivationResult creates a new activation result for given component
// @TODO Hide this from user
func NewActivationResult(componentName string) *ActivationResult {
	return &ActivationResult{
		componentName: componentName,
		Chainable:     common.NewChainable(),
	}
}

// ComponentName getter
func (ar *ActivationResult) ComponentName() string {
	return ar.componentName
}

// Activated getter
func (ar *ActivationResult) Activated() bool {
	return ar.activated
}

// ActivationError getter
func (ar *ActivationResult) ActivationError() error {
	return ar.activationError
}

// ActivationErrorWithComponentName returns activation error enriched with component name
func (ar *ActivationResult) ActivationErrorWithComponentName() error {
	return fmt.Errorf("component %s has activation error: %w", ar.componentName, ar.ActivationError())
}

// Code getter
func (ar *ActivationResult) Code() ActivationResultCode {
	return ar.code
}

// IsError returns true when an activation result has an error
func (ar *ActivationResult) IsError() bool {
	return ar.code == ActivationCodeReturnedError && ar.ActivationError() != nil
}

// IsPanic returns true when an activation result is derived from panic
func (ar *ActivationResult) IsPanic() bool {
	return ar.code == ActivationCodePanicked && ar.ActivationError() != nil
}

// SetActivated setter
func (ar *ActivationResult) SetActivated(activated bool) *ActivationResult {
	ar.activated = activated
	return ar
}

// WithActivationCode setter
func (ar *ActivationResult) WithActivationCode(code ActivationResultCode) *ActivationResult {
	ar.code = code
	return ar
}

// WithActivationError sets the activation result error
func (ar *ActivationResult) WithActivationError(activationError error) *ActivationResult {
	ar.activationError = activationError
	return ar
}

// newActivationResultOK builds a specific activation result
func (c *Component) newActivationResultOK() *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		WithActivationCode(ActivationCodeOK)

}

// newActivationResultNoInput builds a specific activation result
func (c *Component) newActivationResultNoInput() *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(false).
		WithActivationCode(ActivationCodeNoInput)
}

// newActivationResultNoFunction builds a specific activation result
func (c *Component) newActivationResultNoFunction() *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(false).
		WithActivationCode(ActivationCodeNoFunction)
}

// newActivationResultReturnedError builds a specific activation result
func (c *Component) newActivationResultReturnedError(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		WithActivationCode(ActivationCodeReturnedError).
		WithActivationError(fmt.Errorf("component returned an error: %w", err))
}

// newActivationResultPanicked builds a specific activation result
func (c *Component) newActivationResultPanicked(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		WithActivationCode(ActivationCodePanicked).
		WithActivationError(err)
}

func (c *Component) newActivationResultWaitingForInputs(err error) *ActivationResult {
	activationCode := ActivationCodeWaitingForInputsClear
	if errors.Is(err, errWaitingForInputsKeep) {
		activationCode = ActivationCodeWaitingForInputsKeep
	}
	return NewActivationResult(c.Name()).
		SetActivated(true).
		WithActivationCode(activationCode).
		WithActivationError(err)
}

// IsWaitingForInput returns true if the component was waiting for inputs
func IsWaitingForInput(activationResult *ActivationResult) bool {
	return activationResult.Code() == ActivationCodeWaitingForInputsClear ||
		activationResult.Code() == ActivationCodeWaitingForInputsKeep
}

// WantsToKeepInputs returns true if the component wants to keep inputs
func WantsToKeepInputs(activationResult *ActivationResult) bool {
	return activationResult.Code() == ActivationCodeWaitingForInputsKeep
}

// WithErr returns activation result with chain error
func (ar *ActivationResult) WithErr(err error) *ActivationResult {
	ar.SetErr(err)
	return ar
}
