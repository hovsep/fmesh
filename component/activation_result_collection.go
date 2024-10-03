package component

// ActivationResultCollection is a collection
type ActivationResultCollection map[string]*ActivationResult

// NewActivationResultCollection creates empty collection
func NewActivationResultCollection() ActivationResultCollection {
	return make(ActivationResultCollection)
}

// Add adds multiple activation results
func (collection ActivationResultCollection) Add(activationResults ...*ActivationResult) ActivationResultCollection {
	for _, activationResult := range activationResults {
		collection[activationResult.ComponentName()] = activationResult
	}
	return collection
}

// HasErrors tells whether the collection contains at least one activation result with error and respective code
func (collection ActivationResultCollection) HasErrors() bool {
	for _, ar := range collection {
		if ar.IsError() {
			return true
		}
	}
	return false
}

// HasPanics tells whether the collection contains at least one activation result with panic and respective code
func (collection ActivationResultCollection) HasPanics() bool {
	for _, ar := range collection {
		if ar.IsPanic() {
			return true
		}
	}
	return false
}

// HasActivatedComponents tells when at least one component in the cycle has activated
func (collection ActivationResultCollection) HasActivatedComponents() bool {
	for _, ar := range collection {
		if ar.Activated() {
			return true
		}
	}
	return false
}

// ByComponentName returns the activation result of given component
func (collection ActivationResultCollection) ByComponentName(componentName string) *ActivationResult {
	if result, ok := collection[componentName]; ok {
		return result
	}

	return nil
}
