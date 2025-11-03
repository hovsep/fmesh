package port

import (
	"fmt"

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
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		chainableErr: nil,
		ports:        make(Map),
	}
}

// ByName retrieves a specific port from the collection by its name.
// Commonly used to access individual ports from input/output collections.
//
// Example (in activation function):
//
//	data := this.Inputs().ByName("primary").Signals().FirstPayloadOrDefault("")
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

// ByNames retrieves a subset of ports by their names, returning a new collection.
// Useful for operating on a specific group of ports together.
//
// Example (in activation function):
//
//	// Check if specific required inputs have signals
//	if !this.Inputs().ByNames("data", "config").AllHaveSignals() {
//	    return nil // Wait for required inputs
//	}
func (c *Collection) ByNames(names ...string) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}

	selectedPorts := NewCollection()

	for _, name := range names {
		if p, ok := c.ports[name]; ok {
			selectedPorts.With(p)
		}
	}

	return selectedPorts
}

// AnyHasSignals returns true if at least one port in collection has signals.
func (c *Collection) AnyHasSignals() bool {
	return c.AnyMatch(func(p *Port) bool {
		return p.HasSignals()
	})
}

// AllHaveSignals returns true when all ports in the collection have signals.
// Use this to check if all required inputs are ready before processing.
//
// Example (in activation function):
//
//	if !this.Inputs().AllHaveSignals() {
//	    return nil // Wait until all inputs have data
//	}
//	// Process all inputs...
func (c *Collection) AllHaveSignals() bool {
	return c.AllMatch(func(p *Port) bool {
		return p.HasSignals()
	})
}

// PutSignals adds signals to every port in the collection.
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

// ForEach applies an action to each port in the collection and returns it for chaining.
// Use this to perform operations on all ports, such as clearing signals or adding labels.
//
// Example (in activation function):
//
//	// Clear all output ports before writing new data
//	this.Outputs().ForEach(func(p *port.Port) {
//	    p.Clear()
//	})
//
//	// Add labels to all input ports
//	this.Inputs().ForEach(func(p *port.Port) {
//	    p.AddLabel("processed", "true")
//	})
func (c *Collection) ForEach(action func(*Port)) *Collection {
	if c.HasChainableErr() {
		return c
	}
	for _, p := range c.ports {
		action(p)
	}
	return c
}

// Clear removes all ports from the collection.
func (c *Collection) Clear() *Collection {
	if c.HasChainableErr() {
		return c
	}
	c.ports = make(Map)
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
		c.ports[port.Name()] = port
	}

	return c
}

// Without removes ports by name and returns the collection.
func (c *Collection) Without(names ...string) *Collection {
	if c.HasChainableErr() {
		return c
	}

	for _, name := range names {
		delete(c.ports, name)
	}

	return c
}

// WithIndexed creates ports with names like "o1","o2","o3" and so on.
func (c *Collection) WithIndexed(prefix string, startIndex, endIndex int) *Collection {
	if c.HasChainableErr() {
		return c
	}

	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex).All()
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
		signals, err := p.Signals().All()
		if err != nil {
			c.WithChainableErr(err)
			return signal.NewGroup().WithChainableErr(c.ChainableErr())
		}
		group = group.With(signals...)
	}
	return group
}

// All returns all ports as a map.
func (c *Collection) All() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.ports, nil
}

// Any returns any arbitrary port from the collection.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) Any() *Port {
	if c.HasChainableErr() {
		return New("").WithChainableErr(c.ChainableErr())
	}
	if c.IsEmpty() {
		c.WithChainableErr(ErrNoPortsInCollection)
		return New("").WithChainableErr(c.ChainableErr())
	}
	// Get arbitrary port from map (order not guaranteed)
	for _, port := range c.ports {
		return port
	}
	return New("").WithChainableErr(errUnexpectedErrorGettingPort)
}

// AllMatch returns true if all ports match the predicate.
func (c *Collection) AllMatch(predicate Predicate) bool {
	if c.HasChainableErr() {
		return false
	}
	for _, port := range c.ports {
		if !predicate(port) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any port matches the predicate.
func (c *Collection) AnyMatch(predicate Predicate) bool {
	if c.HasChainableErr() {
		return false
	}
	for _, port := range c.ports {
		if predicate(port) {
			return true
		}
	}
	return false
}

// CountMatch returns the number of ports that match the given predicate.
// Use this to count ports with specific characteristics.
//
// Example (in activation function):
//
//	readyCount := this.Inputs().CountMatch(func(p *port.Port) bool {
//	    return p.HasSignals()
//	})
//	this.Logger().Printf("%d inputs ready out of %d", readyCount, this.Inputs().Len())
func (c *Collection) CountMatch(predicate Predicate) int {
	if c.HasChainableErr() {
		return 0
	}
	count := 0
	for _, port := range c.ports {
		if predicate(port) {
			count++
		}
	}
	return count
}

// FindAny returns any arbitrary port that matches the predicate.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) FindAny(predicate Predicate) *Port {
	if c.HasChainableErr() {
		return New("").WithChainableErr(c.ChainableErr())
	}
	for _, port := range c.ports {
		if predicate(port) {
			return port
		}
	}
	c.WithChainableErr(ErrNoPortMatchesPredicate)
	return New("").WithChainableErr(c.ChainableErr())
}

// Filter returns a new collection containing only ports that match the predicate.
// Use this to work with a subset of ports based on specific criteria.
//
// Example (in activation function):
//
//	// Get only ports with signals
//	portsWithData := this.Inputs().Filter(func(p *port.Port) bool {
//	    return p.HasSignals()
//	})
//
//	// Get priority ports
//	priorityPorts := this.Inputs().Filter(func(p *port.Port) bool {
//	    labels, _ := p.Labels().All()
//	    return labels["priority"] == "high"
//	})
func (c *Collection) Filter(predicate Predicate) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	filtered := NewCollection()
	for _, port := range c.ports {
		if predicate(port) {
			filtered = filtered.With(port)
			if filtered.HasChainableErr() {
				return filtered
			}
		}
	}
	return filtered
}

// Map returns a new collection with ports transformed by the mapper function.
func (c *Collection) Map(mapper Mapper) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	mapped := NewCollection()
	for _, port := range c.ports {
		transformedPort := mapper(port)
		if transformedPort != nil {
			mapped = mapped.With(transformedPort)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
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

// ChainableErr returns the chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}

// Len returns the number of ports in a collection.
func (c *Collection) Len() int {
	return len(c.ports)
}

// WithParentComponent adds a parent component to all ports in a collection.
func (c *Collection) WithParentComponent(component ParentComponent) *Collection {
	for _, port := range c.ports {
		port.WithParentComponent(component)
	}
	return c
}

// IsEmpty returns true when there are no ports in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}
