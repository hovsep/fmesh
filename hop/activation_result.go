package hop

// ActivationResult defines the result (possibly an error) of the activation of given component
type ActivationResult struct {
	Activated     bool
	ComponentName string
	Err           error
}
