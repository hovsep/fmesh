package port

// Predicate is a function that tests port matches a condition.
type Predicate func(p *Port) bool

// Mapper transforms a Port into a new Port.
type Mapper func(p *Port) *Port

// Ports is a slice of ports.
type Ports []*Port

// ParentComponent is an interface for components.
type ParentComponent interface {
	Name() string
}
