package component

// ActivationResult defines the result (possibly an error) of the activation of given component in given cycle
type ActivationResult struct {
	activated     bool
	componentName string
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

func (ar ActivationResult) HasError() bool {
	return ar.Error() != nil
}

func (ar ActivationResult) ComponentName() string {
	return ar.componentName
}

func (ar ActivationResult) Activated() bool {
	return ar.activated
}

func (ar ActivationResult) Error() error {
	return ar.err
}

func (c *Component) newActivationResultOK() ActivationResult {
	return ActivationResult{
		activated:     true,
		componentName: c.Name(),
		code:          ActivationCodeOK,
	}
}

func (c *Component) newActivationCodeNoInput() ActivationResult {
	return ActivationResult{
		activated:     false,
		componentName: c.Name(),
		code:          ActivationCodeNoInput,
	}
}

func (c *Component) newActivationCodeNoFunction() ActivationResult {
	return ActivationResult{
		activated:     false,
		componentName: c.Name(),
		code:          ActivationCodeNoFunction,
	}
}

func (c *Component) newActivationCodeWaitingForInput() ActivationResult {
	return ActivationResult{
		activated:     false,
		componentName: c.Name(),
		code:          ActivationCodeWaitingForInput,
	}
}

func (c *Component) newActivationCodeReturnedError(err error) ActivationResult {
	return ActivationResult{
		activated:     true,
		componentName: c.Name(),
		code:          ActivationCodeReturnedError,
		err:           err,
	}
}

func (c *Component) newActivationCodePanicked(err error) ActivationResult {
	return ActivationResult{
		activated:     true,
		componentName: c.Name(),
		code:          ActivationCodePanicked,
		err:           err,
	}
}
