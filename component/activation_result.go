package component

// ActivationResult defines the result (possibly an error) of the activation of given component in given cycle
type ActivationResult struct {
	componentName string
	activated     bool
	code          ActivationCode
	err           error
}

// ActivationCode denotes a specific info about how a component been activated or why not activated at all
type ActivationCode int

const (
	// ActivationCodeOK ...: component is activated and did not return any errors
	ActivationCodeOK ActivationCode = iota

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

func New(componentName string) *ActivationResult {
	return &ActivationResult{
		componentName: componentName,
	}
}

func (ar *ActivationResult) SetActivated(activated bool) *ActivationResult {
	ar.activated = activated
	return ar
}

func (ar *ActivationResult) WithActivationCode(code ActivationCode) *ActivationResult {
	ar.code = code
	return ar
}

func (ar *ActivationResult) WithError(err error) *ActivationResult {
	ar.err = err
	return ar
}

func (c *Component) newActivationResultOK() *ActivationResult {
	return New(c.Name()).SetActivated(true).WithActivationCode(ActivationCodeOK)
}

func (c *Component) newActivationCodeNoInput() *ActivationResult {
	return New(c.Name()).SetActivated(false).WithActivationCode(ActivationCodeNoInput)
}

func (c *Component) newActivationCodeNoFunction() *ActivationResult {
	return New(c.Name()).SetActivated(false).WithActivationCode(ActivationCodeNoFunction)
}

func (c *Component) newActivationCodeWaitingForInput() *ActivationResult {
	return New(c.Name()).SetActivated(false).WithActivationCode(ActivationCodeWaitingForInput)
}

func (c *Component) newActivationCodeReturnedError(err error) *ActivationResult {
	return New(c.Name()).SetActivated(true).WithActivationCode(ActivationCodeReturnedError).WithError(err)
}

func (c *Component) newActivationCodePanicked(err error) *ActivationResult {
	return New(c.Name()).SetActivated(true).WithActivationCode(ActivationCodePanicked).WithError(err)
}

func (ar *ActivationResult) ComponentName() string {
	return ar.componentName
}

func (ar *ActivationResult) Activated() bool {
	return ar.activated
}

func (ar *ActivationResult) Error() error {
	return ar.err
}

func (ar *ActivationResult) HasError() bool {
	return ar.code == ActivationCodeReturnedError && ar.Error() != nil
}

func (ar *ActivationResult) HasPanic() bool {
	return ar.code == ActivationCodePanicked && ar.Error() != nil
}
