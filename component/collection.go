package component

// Collection is a collection of components with useful methods
type Collection map[string]*Component

// NewComponentCollection creates empty collection
func NewComponentCollection() Collection {
	return make(Collection)
}

// ByName returns a component by its name
func (collection Collection) ByName(name string) *Component {
	return collection[name]
}

// Add adds components to existing collection
func (collection Collection) Add(components ...*Component) Collection {
	for _, component := range components {
		collection[component.Name()] = component
	}
	return collection
}
