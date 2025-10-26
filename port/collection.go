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
func (c *Collection) ByName(name string) *Port {
	if c.HasChainableErr() {
		return New("").WithChainableErr(c.ChainableErr())
	}
	port, ok := c.ports[name]
	if !ok {
		c.WithChainableErr(fmt.Errorf("%w, port name: %s", ErrPortNotFoundInCollection, name))
		return New("").WithChainableErr(c.ChainableErr())
	}
	return port
}

// ByNames returns multiple ports by their names.
func (c *Collection) ByNames(names ...string) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}

	// Preserve c config
	selectedPorts := NewCollection().WithDefaultLabels(c.defaultLabels)

	for _, name := range names {
		if p, ok := c.ports[name]; ok {
			selectedPorts.With(p)
		}
	}

	return selectedPorts
}

// AnyHasSignals returns true if at least one port in collection has signals.
func (c *Collection) AnyHasSignals() bool {
	if c.HasChainableErr() {
		return false
	}

	for _, p := range c.ports {
		if p.HasSignals() {
			return true
		}
	}

	return false
}

// AllHaveSignals returns true when all ports in collection have signals.
func (c *Collection) AllHaveSignals() bool {
	if c.HasChainableErr() {
		return false
	}

	for _, p := range c.ports {
		if !p.HasSignals() {
			return false
		}
	}

	return true
}

// PutSignals adds buffer to every port in collection.
func (c *Collection) PutSignals(signals ...*signal.Signal) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}

	for _, p := range c.ports {
		p.PutSignals(signals...)
		if p.HasChainableErr() {
			return c.WithChainableErr(p.ChainableErr())
		}
	}

	return c
}

// Clear clears all ports in collection.
func (c *Collection) Clear() *Collection {
	for _, p := range c.ports {
		p.Clear()

		if p.HasChainableErr() {
			return c.WithChainableErr(p.ChainableErr())
		}
	}
	return c
}

// Flush flushes all ports in collection.
func (c *Collection) Flush() *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}

	for _, p := range c.ports {
		p = p.Flush()

		if p.HasChainableErr() {
			return c.WithChainableErr(p.ChainableErr())
		}
	}
	return c
}

// PipeTo creates pipes from each port in collection to given destination ports.
func (c *Collection) PipeTo(destPorts ...*Port) *Collection {
	for _, p := range c.ports {
		p = p.PipeTo(destPorts...)

		if p.HasChainableErr() {
			return c.WithChainableErr(p.ChainableErr())
		}
	}

	return c
}

// With adds ports to collection and returns it.
func (c *Collection) With(ports ...*Port) *Collection {
	if c.HasChainableErr() {
		return c
	}

	for _, port := range ports {
		if port.HasChainableErr() {
			return c.WithChainableErr(port.ChainableErr())
		}
		port.labels.WithMany(c.defaultLabels)
		c.ports[port.Name()] = port
	}

	return c
}

// WithIndexed creates ports with names like "o1","o2","o3" and so on.
func (c *Collection) WithIndexed(prefix string, startIndex, endIndex int) *Collection {
	if c.HasChainableErr() {
		return c
	}

	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex).Ports()
	if err != nil {
		c.WithChainableErr(err)
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	return c.With(indexedPorts...)
}

// Signals returns all signals of all ports in the collection.
func (c *Collection) Signals() *signal.Group {
	if c.HasChainableErr() {
		return signal.NewGroup().WithChainableErr(c.ChainableErr())
	}

	group := signal.NewGroup()
	for _, p := range c.ports {
		signals, err := p.Buffer().Signals()
		if err != nil {
			c.WithChainableErr(err)
			return signal.NewGroup().WithChainableErr(c.ChainableErr())
		}
		group = group.With(signals...)
	}
	return group
}

// Ports getter
// @TODO:maybe better to hide all errors within chainable and ask user to check error ?
func (c *Collection) Ports() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.ports, nil
}

// PortsOrNil returns ports or nil in case of any error.
func (c *Collection) PortsOrNil() Map {
	return c.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error.
func (c *Collection) PortsOrDefault(defaultPorts Map) Map {
	if c.HasChainableErr() {
		return defaultPorts
	}

	ports, err := c.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// WithChainableErr sets a chainable error and returns the collection.
func (c *Collection) WithChainableErr(err error) *Collection {
	c.chainableErr = err
	return c
}

// HasChainableErr returns true when a chainable error is set.
func (c *Collection) HasChainableErr() bool {
	return c.chainableErr != nil
}

// ChainableErr returns chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}

// Len returns the number of ports in a collection.
func (c *Collection) Len() int {
	return len(c.ports)
}

// WithDefaultLabels adds default labels to all ports in collection.
func (c *Collection) WithDefaultLabels(labels labels.Map) *Collection {
	c.defaultLabels = labels
	return c
}
