package component

import "fmt"

// ActivationResult defines the result (possibly an error) of the activation of given component in given cycle
type ActivationResult struct {
	componentName string
	activated     bool
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

// newActivationResultOK builds a specific activation result
func (c *Component) newActivationResultOK() *ActivationResult {
	return NewActivationResult(c.Name()).SetActivated(true).WithActivationCode(ActivationCodeOK)
}

// newActivationCodeNoInput builds a specific activation result
func (c *Component) newActivationCodeNoInput() *ActivationResult {
	return NewActivationResult(c.Name()).SetActivated(false).WithActivationCode(ActivationCodeNoInput)
}

// newActivationCodeNoFunction builds a specific activation result
func (c *Component) newActivationCodeNoFunction() *ActivationResult {
	return NewActivationResult(c.Name()).SetActivated(false).WithActivationCode(ActivationCodeNoFunction)
}

// newActivationCodeWaitingForInput builds a specific activation result
func (c *Component) newActivationCodeWaitingForInput() *ActivationResult {
	return NewActivationResult(c.Name()).SetActivated(false).WithActivationCode(ActivationCodeWaitingForInput)
}

// newActivationCodeReturnedError builds a specific activation result
func (c *Component) newActivationCodeReturnedError(err error) *ActivationResult {
	return NewActivationResult(c.Name()).SetActivated(true).WithActivationCode(ActivationCodeReturnedError).WithError(fmt.Errorf("component returned an error: %w", err))
}

// newActivationCodePanicked builds a specific activation result
func (c *Component) newActivationCodePanicked(err error) *ActivationResult {
	return NewActivationResult(c.Name()).SetActivated(true).WithActivationCode(ActivationCodePanicked).WithError(err)
}
