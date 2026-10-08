package port

// Predicate is a function that tests port matches a condition.
type Predicate func(p *Port) bool

// ParentComponent is an interface for components.
type ParentComponent interface {
	Name() string
}
