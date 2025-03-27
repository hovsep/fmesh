package component

import "sync"

// ActivationResultCollection is a collection
type ActivationResultCollection struct {
	mu sync.Mutex
	activationResults map[string]*ActivationResult
}

// NewActivationResultCollection creates empty collection
func NewActivationResultCollection() *ActivationResultCollection {
	return &ActivationResultCollection{
		activationResults: make(map[string]*ActivationResult),
	}
}

// Add adds multiple activation results
func (collection *ActivationResultCollection) Add(activationResults ...*ActivationResult) *ActivationResultCollection {
	collection.mu.Lock()
	defer collection.mu.Unlock()

	for _, activationResult := range activationResults {
		collection.activationResults[activationResult.ComponentName()] = activationResult
	}
	return collection
}

// HasErrors tells whether the collection contains at least one activation result with error and respective code
func (collection *ActivationResultCollection) HasErrors() bool {
	for _, ar := range collection.activationResults {
		if ar.IsError() {
			return true
		}
	}
	return false
}

// HasPanics tells whether the collection contains at least one activation result with panic and respective code
func (collection *ActivationResultCollection) HasPanics() bool {
	for _, ar := range collection.activationResults {
		if ar.IsPanic() {
			return true
		}
	}
	return false
}

// HasActivatedComponents tells when at least one component in the cycle has activated
func (collection *ActivationResultCollection) HasActivatedComponents() bool {
	for _, ar := range collection.activationResults {
		if ar.Activated() {
			return true
		}
	}
	return false
}

// ByComponentName returns the activation result of given component
func (collection *ActivationResultCollection) ByComponentName(componentName string) *ActivationResult {
	if result, ok := collection.activationResults[componentName]; ok {
		return result
	}

	return nil
}

// All returns all activation results
func (collection *ActivationResultCollection) All() map[string]*ActivationResult{
	return collection.activationResults
}

// Len returns the number of activation results in the collection
func (collection *ActivationResultCollection) Len() int {
	return len(collection.activationResults)
}
