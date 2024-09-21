package component

import (
	"errors"
	"fmt"
)

// ActivationResult defines the result (possibly an error) of the activation of given component in given cycle
type ActivationResult struct {
	componentName string
	activated     bool
	stateBefore   *StateSnapshot //Contains the info about length of input ports during the activation (required for correct i2i piping)
	stateAfter    *StateSnapshot
	code          ActivationResultCode
	err           error
}

// ActivationResultCode denotes a specific info about how a component been activated or why not activated at all
type ActivationResultCode int

const (
	// ActivationCodeOK ...: component is activated and did not return any errors
	ActivationCodeOK ActivationResultCode = iota

	// ActivationCodeNoInput ...: component is not activated because it has no input set
	ActivationCodeNoInput

	// ActivationCodeNoFunction ...: component activation function is not set, so we can not activate it
	ActivationCodeNoFunction

	// ActivationCodeWaitingForInput ...: component is waiting for more inputs on some ports
	ActivationCodeWaitingForInput

	// ActivationCodeReturnedError ...: component is activated, but returned an error
	ActivationCodeReturnedError

	// ActivationCodePanicked ...: component is activated, but panicked
	ActivationCodePanicked
)

// NewActivationResult creates a new activation result for given component
// @TODO Hide this from user
func NewActivationResult(componentName string) *ActivationResult {
	return &ActivationResult{
		componentName: componentName,
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

// Error getter
func (ar *ActivationResult) Error() error {
	return ar.err
}

// Code getter
func (ar *ActivationResult) Code() ActivationResultCode {
	return ar.code
}

// HasError returns true when activation result has an error
func (ar *ActivationResult) HasError() bool {
	return ar.code == ActivationCodeReturnedError && ar.Error() != nil
}

// HasPanic returns true when activation result is derived from panic
func (ar *ActivationResult) HasPanic() bool {
	return ar.code == ActivationCodePanicked && ar.Error() != nil
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

// WithError setter
func (ar *ActivationResult) WithError(err error) *ActivationResult {
	ar.err = err
	return ar
}

func (ar *ActivationResult) WithStateBefore(snapshot *StateSnapshot) *ActivationResult {
	ar.stateBefore = snapshot
	return ar
}

func (ar *ActivationResult) StateBefore() *StateSnapshot {
	return ar.stateBefore
}

func (ar *ActivationResult) WithStateAfter(snapshot *StateSnapshot) *ActivationResult {
	ar.stateAfter = snapshot
	return ar
}

func (ar *ActivationResult) StateAfter() *StateSnapshot {
	return ar.stateAfter
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

// newActivationResultWaitingForInput builds a specific activation result
func (c *Component) newActivationResultWaitingForInput() *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(false).
		WithActivationCode(ActivationCodeWaitingForInput)
}

// newActivationResultReturnedError builds a specific activation result
func (c *Component) newActivationResultReturnedError(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		WithActivationCode(ActivationCodeReturnedError).
		WithError(fmt.Errorf("component returned an error: %w", err))
}

// newActivationResultPanicked builds a specific activation result
func (c *Component) newActivationResultPanicked(err error) *ActivationResult {
	return NewActivationResult(c.Name()).
		SetActivated(true).
		WithActivationCode(ActivationCodePanicked).
		WithError(err)
}

// isWaitingForInput tells whether component is waiting for specific inputs
func (c *Component) isWaitingForInput(activationResult *ActivationResult) bool {
	return activationResult.HasError() && errors.Is(activationResult.Error(), errWaitingForInputs)
}

// WantsToKeepInputs tells whether component wants to keep signals on input ports for the next cycle
func (c *Component) WantsToKeepInputs(activationResult *ActivationResult) bool {
	return c.isWaitingForInput(activationResult) && errors.Is(activationResult.Error(), errWaitingForInputsKeep)
}
