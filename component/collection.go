package component

// ComponentCollection is a collection of components with useful methods
type ComponentCollection map[string]*Component

// NewComponentCollection creates empty collection
func NewComponentCollection() ComponentCollection {
	return make(ComponentCollection)
}

// ByName returns a component by its name
func (collection ComponentCollection) ByName(name string) *Component {
	return collection[name]
}

// Add adds new components to existing collection
func (collection ComponentCollection) Add(newComponents ...*Component) ComponentCollection {
	for _, component := range newComponents {
		collection[component.Name()] = component
	}
	return collection
}
