package component

// Predicate is a function that tests whether a Component matches a condition.
type Predicate func(component *Component) bool

// Mapper transforms a Component into a new Component.
type Mapper func(component *Component) *Component

// Components is a slice of components for type safety and method attachment.
type Components []*Component

// ActivationResultPredicate is a function that tests whether an ActivationResult matches a condition.
type ActivationResultPredicate func(result *ActivationResult) bool

// ActivationResultMapper transforms an ActivationResult into a new ActivationResult.
type ActivationResultMapper func(result *ActivationResult) *ActivationResult

// ActivationResults is a slice of activation results for type safety and method attachment.
type ActivationResults []*ActivationResult

// ParentMesh is an interface for a parent mesh.
type ParentMesh interface {
	Name() string
}

// ActivationFunc is the activation function of a component.
type ActivationFunc func(this *Component) error
