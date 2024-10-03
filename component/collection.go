package component

// Collection is a collection of components with useful methods
type Collection map[string]*Component

// NewCollection creates empty collection
func NewCollection() Collection {
	return make(Collection)
}

// ByName returns a component by its name
func (collection Collection) ByName(name string) *Component {
	return collection[name]
}

// With adds components and returns the collection
func (collection Collection) With(components ...*Component) Collection {
	for _, component := range components {
		collection[component.Name()] = component
	}
	return collection
}
