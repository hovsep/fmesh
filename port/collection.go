package port

import (
	"fmt"

	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/signal"
)

// Map is a map of ports.
type Map map[string]*Port

// Collection is a port collection.
// indexed by name; hence it cannot carry
// 2 ports with the same name. Optimized for lookups.
type Collection struct {
	*common.Chainable
	ports Map
	// Labels added by default to each port in collection
	defaultLabels common.LabelsCollection
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		Chainable:     common.NewChainable(),
		ports:         make(Map),
		defaultLabels: common.LabelsCollection{},
	}
}

// ByName returns a port by its name.
func (collection *Collection) ByName(name string) *Port {
	if collection.HasErr() {
		return New("").WithErr(collection.Err())
	}
	port, ok := collection.ports[name]
	if !ok {
		collection.SetErr(fmt.Errorf("%w, port name: %s", ErrPortNotFoundInCollection, name))
		return New("").WithErr(collection.Err())
	}
	return port
}

// ByNames returns multiple ports by their names.
func (collection *Collection) ByNames(names ...string) *Collection {
	if collection.HasErr() {
		return NewCollection().WithErr(collection.Err())
	}

	// Preserve collection config
	selectedPorts := NewCollection().WithDefaultLabels(collection.defaultLabels)

	for _, name := range names {
		if p, ok := collection.ports[name]; ok {
			selectedPorts.With(p)
		}
	}

	return selectedPorts
}

// AnyHasSignals returns true if at least one port in collection has signals.
func (collection *Collection) AnyHasSignals() bool {
	if collection.HasErr() {
		return false
	}

	for _, p := range collection.ports {
		if p.HasSignals() {
			return true
		}
	}

	return false
}

// AllHaveSignals returns true when all ports in collection have signals.
func (collection *Collection) AllHaveSignals() bool {
	if collection.HasErr() {
		return false
	}

	for _, p := range collection.ports {
		if !p.HasSignals() {
			return false
		}
	}

	return true
}

// PutSignals adds buffer to every port in collection.
func (collection *Collection) PutSignals(signals ...*signal.Signal) *Collection {
	if collection.HasErr() {
		return NewCollection().WithErr(collection.Err())
	}

	for _, p := range collection.ports {
		p.PutSignals(signals...)
		if p.HasErr() {
			return collection.WithErr(p.Err())
		}
	}

	return collection
}

// Clear clears all ports in collection.
func (collection *Collection) Clear() *Collection {
	for _, p := range collection.ports {
		p.Clear()

		if p.HasErr() {
			return collection.WithErr(p.Err())
		}
	}
	return collection
}

// Flush flushes all ports in collection.
func (collection *Collection) Flush() *Collection {
	if collection.HasErr() {
		return NewCollection().WithErr(collection.Err())
	}

	for _, p := range collection.ports {
		p = p.Flush()

		if p.HasErr() {
			return collection.WithErr(p.Err())
		}
	}
	return collection
}

// PipeTo creates pipes from each port in collection to given destination ports.
func (collection *Collection) PipeTo(destPorts ...*Port) *Collection {
	for _, p := range collection.ports {
		p = p.PipeTo(destPorts...)

		if p.HasErr() {
			return collection.WithErr(p.Err())
		}
	}

	return collection
}

// With adds ports to collection and returns it.
func (collection *Collection) With(ports ...*Port) *Collection {
	if collection.HasErr() {
		return collection
	}

	for _, port := range ports {
		if port.HasErr() {
			return collection.WithErr(port.Err())
		}
		port.AddLabels(collection.defaultLabels)
		collection.ports[port.Name()] = port
	}

	return collection
}

// WithIndexed creates ports with names like "o1","o2","o3" and so on.
func (collection *Collection) WithIndexed(prefix string, startIndex, endIndex int) *Collection {
	if collection.HasErr() {
		return collection
	}

	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex).Ports()
	if err != nil {
		collection.SetErr(err)
		return NewCollection().WithErr(collection.Err())
	}
	return collection.With(indexedPorts...)
}

// Signals returns all signals of all ports in the collection.
func (collection *Collection) Signals() *signal.Group {
	if collection.HasErr() {
		return signal.NewGroup().WithErr(collection.Err())
	}

	group := signal.NewGroup()
	for _, p := range collection.ports {
		signals, err := p.Buffer().Signals()
		if err != nil {
			collection.SetErr(err)
			return signal.NewGroup().WithErr(collection.Err())
		}
		group = group.With(signals...)
	}
	return group
}

// Ports getter
// @TODO:maybe better to hide all errors within chainable and ask user to check error ?
func (collection *Collection) Ports() (Map, error) {
	if collection.HasErr() {
		return nil, collection.Err()
	}
	return collection.ports, nil
}

// PortsOrNil returns ports or nil in case of any error.
func (collection *Collection) PortsOrNil() Map {
	return collection.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error.
func (collection *Collection) PortsOrDefault(defaultPorts Map) Map {
	if collection.HasErr() {
		return defaultPorts
	}

	ports, err := collection.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// WithErr returns group with error.
func (collection *Collection) WithErr(err error) *Collection {
	collection.SetErr(err)
	return collection
}

// Len returns number of ports in collection.
func (collection *Collection) Len() int {
	return len(collection.ports)
}

// WithDefaultLabels adds default labels to all ports in collection.
func (collection *Collection) WithDefaultLabels(labels common.LabelsCollection) *Collection {
	collection.defaultLabels = labels
	return collection
}
