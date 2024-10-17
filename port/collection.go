package port

import (
	"errors"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/signal"
)

// Collection is a port collection
// indexed by name, hence it can not carry
// 2 ports with same name. Optimized for lookups
type Collection struct {
	*common.Chainable
	ports map[string]*Port
}

// NewCollection creates empty collection
func NewCollection() *Collection {
	return &Collection{
		Chainable: common.NewChainable(),
		ports:     make(map[string]*Port),
	}
}

// ByName returns a port by its name
func (collection *Collection) ByName(name string) *Port {
	if collection.HasChainError() {
		return nil
	}
	port, ok := collection.ports[name]
	if !ok {
		collection.SetChainError(errors.New("port not found"))
		return nil
	}
	return port
}

// ByNames returns multiple ports by their names
func (collection *Collection) ByNames(names ...string) *Collection {
	if collection.HasChainError() {
		return collection
	}

	selectedPorts := NewCollection()

	for _, name := range names {
		if p, ok := collection.ports[name]; ok {
			selectedPorts.With(p)
		}
	}

	return selectedPorts
}

// AnyHasSignals returns true if at least one port in collection has signals
func (collection *Collection) AnyHasSignals() bool {
	if collection.HasChainError() {
		return false
	}

	for _, p := range collection.ports {
		if p.HasSignals() {
			return true
		}
	}

	return false
}

// AllHaveSignals returns true when all ports in collection have signals
func (collection *Collection) AllHaveSignals() bool {
	if collection.HasChainError() {
		return false
	}

	for _, p := range collection.ports {
		if !p.HasSignals() {
			return false
		}
	}

	return true
}

// PutSignals adds buffer to every port in collection
// @TODO: return collection
func (collection *Collection) PutSignals(signals ...*signal.Signal) *Collection {
	if collection.HasChainError() {
		return collection
	}

	for _, p := range collection.ports {
		p.PutSignals(signals...)
		if p.HasChainError() {
			return collection.WithChainError(p.ChainError())
		}
	}

	return collection
}

// Clear clears all ports in collection
func (collection *Collection) Clear() *Collection {
	for _, p := range collection.ports {
		p.Clear()

		if p.HasChainError() {
			return collection.WithChainError(p.ChainError())
		}
	}
	return collection
}

// Flush flushes all ports in collection
func (collection *Collection) Flush() *Collection {
	if collection.HasChainError() {
		return collection
	}

	for _, p := range collection.ports {
		p.Flush()

		if p.HasChainError() {
			return collection.WithChainError(p.ChainError())
		}
	}
	return collection
}

// PipeTo creates pipes from each port in collection to given destination ports
func (collection *Collection) PipeTo(destPorts ...*Port) *Collection {
	for _, p := range collection.ports {
		p.PipeTo(destPorts...)

		if p.HasChainError() {
			return collection.WithChainError(p.ChainError())
		}
	}

	return collection
}

// With adds ports to collection and returns it
func (collection *Collection) With(ports ...*Port) *Collection {
	if collection.HasChainError() {
		return collection
	}

	for _, port := range ports {
		collection.ports[port.Name()] = port

		if port.HasChainError() {
			return collection.WithChainError(port.ChainError())
		}
	}

	return collection
}

// WithIndexed creates ports with names like "o1","o2","o3" and so on
func (collection *Collection) WithIndexed(prefix string, startIndex int, endIndex int) *Collection {
	if collection.HasChainError() {
		return collection
	}

	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex).Ports()
	if err != nil {
		return collection.WithChainError(err)
	}
	return collection.With(indexedPorts...)
}

// Signals returns all signals of all ports in the collection
func (collection *Collection) Signals() *signal.Group {
	if collection.HasChainError() {
		return signal.NewGroup().WithChainError(collection.ChainError())
	}

	group := signal.NewGroup()
	for _, p := range collection.ports {
		signals, err := p.Buffer().Signals()
		if err != nil {
			return group.WithChainError(err)
		}
		group = group.With(signals...)
	}
	return group
}

// withPorts sets ports
func (collection *Collection) withPorts(ports map[string]*Port) *Collection {
	if collection.HasChainError() {
		return collection
	}

	collection.ports = ports
	return collection
}

// Ports getter
// @TODO:maybe better to hide all errors within chainable and ask user to check error ?
func (collection *Collection) Ports() (map[string]*Port, error) {
	if collection.HasChainError() {
		return nil, collection.ChainError()
	}
	return collection.ports, nil
}

// PortsOrNil returns ports or nil in case of any error
func (collection *Collection) PortsOrNil() map[string]*Port {
	return collection.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error
func (collection *Collection) PortsOrDefault(defaultPorts map[string]*Port) map[string]*Port {
	if collection.HasChainError() {
		return defaultPorts
	}

	ports, err := collection.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// WithChainError returns group with error
func (collection *Collection) WithChainError(err error) *Collection {
	collection.SetChainError(err)
	return collection
}

// Len returns number of ports in collection
func (collection *Collection) Len() int {
	return len(collection.ports)
}
