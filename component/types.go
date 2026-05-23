package component

// Predicate is a function that tests whether a Component matches a condition.
type Predicate func(component *Component) bool

// Mapper transforms a Component into a new Component.
type Mapper func(component *Component) *Component

// Components is a slice of components for type safety and method attachment.
type Components []*Component

// ResultPredicate is a function that tests whether an ActivationResult matches a condition.
type ResultPredicate func(result *ActivationResult) bool

// ResultMapper transforms an ActivationResult into a new ActivationResult.
type ResultMapper func(result *ActivationResult) *ActivationResult

// ActivationResults is a slice of activation results for type safety and method attachment.
type ActivationResults []*ActivationResult

// ParentMesh is an interface for a parent mesh.
type ParentMesh interface {
	Name() string
}

// ActivationFunc is the activation function of a component.
type ActivationFunc func(this *Component) error

// Option is a functional option for configuring a component during construction.
type Option func(*Component) error
