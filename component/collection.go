package component

import (
	"errors"
	"github.com/hovsep/fmesh/common"
)

type ComponentsMap map[string]*Component

// Collection is a collection of components with useful methods
type Collection struct {
	*common.Chainable
	components ComponentsMap
}

// NewCollection creates empty collection
func NewCollection() *Collection {
	return &Collection{
		Chainable:  common.NewChainable(),
		components: make(ComponentsMap),
	}
}

// ByName returns a component by its name
func (collection *Collection) ByName(name string) *Component {
	if collection.HasChainError() {
		return nil
	}

	component, ok := collection.components[name]

	if !ok {
		collection.SetChainError(errors.New("component not found"))
		return nil
	}

	return component
}

// With adds components and returns the collection
func (collection *Collection) With(components ...*Component) *Collection {
	if collection.HasChainError() {
		return collection
	}

	for _, component := range components {
		collection.components[component.Name()] = component

		if component.HasChainError() {
			return collection.WithChainError(component.ChainError())
		}
	}

	return collection
}

// WithChainError returns group with error
func (collection *Collection) WithChainError(err error) *Collection {
	collection.SetChainError(err)
	return collection
}

// Len returns number of ports in collection
func (collection *Collection) Len() int {
	return len(collection.components)
}

func (collection *Collection) Components() (ComponentsMap, error) {
	if collection.HasChainError() {
		return nil, collection.ChainError()
	}
	return collection.components, nil
}
