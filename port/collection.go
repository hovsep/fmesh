package port

import (
	"fmt"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
)

// Map is a map of ports.
type Map map[string]*Port

// Collection is a port collection.
// indexed by name; hence it cannot carry
// 2 ports with the same name. Optimized for lookups.
type Collection struct {
	chainableErr error
	ports        Map
	// Labels added by default to each port in collection
	defaultLabels labels.Map
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		chainableErr:  nil,
		ports:         make(Map),
		defaultLabels: labels.Map{},
	}
}

// ByName returns a port by its name.
func (collection *Collection) ByName(name string) *Port {
	if collection.HasChainableErr() {
		return New("").WithChainableErr(collection.ChainableErr())
	}
	port, ok := collection.ports[name]
	if !ok {
		collection.WithChainableErr(fmt.Errorf("%w, port name: %s", ErrPortNotFoundInCollection, name))
		return New("").WithChainableErr(collection.ChainableErr())
	}
	return port
}

// ByNames returns multiple ports by their names.
func (collection *Collection) ByNames(names ...string) *Collection {
	if collection.HasChainableErr() {
		return NewCollection().WithChainableErr(collection.ChainableErr())
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
	if collection.HasChainableErr() {
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
	if collection.HasChainableErr() {
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
	if collection.HasChainableErr() {
		return NewCollection().WithChainableErr(collection.ChainableErr())
	}

	for _, p := range collection.ports {
		p.PutSignals(signals...)
		if p.HasChainableErr() {
			return collection.WithChainableErr(p.ChainableErr())
		}
	}

	return collection
}

// Clear clears all ports in collection.
func (collection *Collection) Clear() *Collection {
	for _, p := range collection.ports {
		p.Clear()

		if p.HasChainableErr() {
			return collection.WithChainableErr(p.ChainableErr())
		}
	}
	return collection
}

// Flush flushes all ports in collection.
func (collection *Collection) Flush() *Collection {
	if collection.HasChainableErr() {
		return NewCollection().WithChainableErr(collection.ChainableErr())
	}

	for _, p := range collection.ports {
		p = p.Flush()

		if p.HasChainableErr() {
			return collection.WithChainableErr(p.ChainableErr())
		}
	}
	return collection
}

// PipeTo creates pipes from each port in collection to given destination ports.
func (collection *Collection) PipeTo(destPorts ...*Port) *Collection {
	for _, p := range collection.ports {
		p = p.PipeTo(destPorts...)

		if p.HasChainableErr() {
			return collection.WithChainableErr(p.ChainableErr())
		}
	}

	return collection
}

// With adds ports to collection and returns it.
func (collection *Collection) With(ports ...*Port) *Collection {
	if collection.HasChainableErr() {
		return collection
	}

	for _, port := range ports {
		if port.HasChainableErr() {
			return collection.WithChainableErr(port.ChainableErr())
		}
		port.labels.WithMany(collection.defaultLabels)
		collection.ports[port.Name()] = port
	}

	return collection
}

// WithIndexed creates ports with names like "o1","o2","o3" and so on.
func (collection *Collection) WithIndexed(prefix string, startIndex, endIndex int) *Collection {
	if collection.HasChainableErr() {
		return collection
	}

	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex).Ports()
	if err != nil {
		collection.WithChainableErr(err)
		return NewCollection().WithChainableErr(collection.ChainableErr())
	}
	return collection.With(indexedPorts...)
}

// Signals returns all signals of all ports in the collection.
func (collection *Collection) Signals() *signal.Group {
	if collection.HasChainableErr() {
		return signal.NewGroup().WithChainableErr(collection.ChainableErr())
	}

	group := signal.NewGroup()
	for _, p := range collection.ports {
		signals, err := p.Buffer().Signals()
		if err != nil {
			collection.WithChainableErr(err)
			return signal.NewGroup().WithChainableErr(collection.ChainableErr())
		}
		group = group.With(signals...)
	}
	return group
}

// Ports getter
// @TODO:maybe better to hide all errors within chainable and ask user to check error ?
func (collection *Collection) Ports() (Map, error) {
	if collection.HasChainableErr() {
		return nil, collection.ChainableErr()
	}
	return collection.ports, nil
}

// PortsOrNil returns ports or nil in case of any error.
func (collection *Collection) PortsOrNil() Map {
	return collection.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error.
func (collection *Collection) PortsOrDefault(defaultPorts Map) Map {
	if collection.HasChainableErr() {
		return defaultPorts
	}

	ports, err := collection.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// WithChainableErr sets a chainable error and returns the collection.
func (collection *Collection) WithChainableErr(err error) *Collection {
	collection.chainableErr = err
	return collection
}

// HasChainableErr returns true when a chainable error is set.
func (collection *Collection) HasChainableErr() bool {
	return collection.chainableErr != nil
}

// ChainableErr returns chainable error.
func (collection *Collection) ChainableErr() error {
	return collection.chainableErr
}

// Len returns the number of ports in a collection.
func (collection *Collection) Len() int {
	return len(collection.ports)
}

// WithDefaultLabels adds default labels to all ports in collection.
func (collection *Collection) WithDefaultLabels(labels labels.Map) *Collection {
	collection.defaultLabels = labels
	return collection
}
