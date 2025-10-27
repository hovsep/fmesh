package component

// Predicate is a function that tests whether a Component matches a condition.
type Predicate func(component *Component) bool

// Mapper transforms a Component into a new Component.
type Mapper func(component *Component) *Component

// Components is a slice of components for type safety and method attachment.
type Components []*Component

// ParentMesh is an interface for a parent mesh.
type ParentMesh interface {
	Name() string
}

// ActivationFunc is the activation function of a component.
type ActivationFunc func(this *Component) error
