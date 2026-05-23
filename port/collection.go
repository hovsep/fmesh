package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Map is a map of ports.
type Map map[string]*Port

// Collection is a port collection.
// indexed by name; hence it cannot carry
// 2 ports with the same name. Optimized for lookups.
// @TODO: do we need this type if now it holds only 1 field?
type Collection struct {
	ports Map
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		ports: make(Map),
	}
}

// ByName retrieves a specific port from the collection by its name.
// Returns nil if not found.
//
// Example (in activation function):
//
//	if port := this.Inputs().ByName("primary"); port != nil {
//	    data := port.Signals().FirstPayloadOrDefault("")
//	}
func (c *Collection) ByName(name string) *Port {
	return c.ports[name]
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
	selected := NewCollection()
	for _, name := range names {
		if p, ok := c.ports[name]; ok {
			selected.Add(p)
		}
	}
	return selected
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
	return c.Every(func(p *Port) bool {
		return p.HasSignals()
	})
}

// PutSignals adds signals to every port in the collection.
// Stops and returns the first error encountered.
func (c *Collection) PutSignals(signals ...*signal.Signal) error {
	for _, p := range c.ports {
		if err := p.PutSignals(signals...); err != nil {
			return err
		}
	}
	return nil
}

// ForEach applies an action to each port in the collection.
// Returns the first error encountered.
//
// Example (in activation function):
//
//	// Add labels to all input ports
//	this.Inputs().ForEach(func(p *port.Port) error {
//	    p.AddLabel("processed", "true")
//	    return nil
//	})
func (c *Collection) ForEach(action func(*Port) error) error {
	for _, p := range c.ports {
		if err := action(p); err != nil {
			return err
		}
	}
	return nil
}

// Flush flushes all ports in a collection.
// Stops and returns the first error encountered.
func (c *Collection) Flush() error {
	for _, p := range c.ports {
		if err := p.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// PipeTo creates pipes from each port in a collection to given destination ports.
// Stops and returns the first error encountered.
func (c *Collection) PipeTo(destPorts ...*Port) error {
	for _, p := range c.ports {
		if err := p.PipeTo(destPorts...); err != nil {
			return err
		}
	}
	return nil
}

// Add adds ports to a collection and returns it. Overwrites on name conflict.
func (c *Collection) Add(ports ...*Port) *Collection {
	for _, port := range ports {
		c.ports[port.Name()] = port
	}
	return c
}

// Without removes ports by name and returns the collection.
func (c *Collection) Without(names ...string) *Collection {
	for _, name := range names {
		delete(c.ports, name)
	}
	return c
}

// AddIndexed creates ports with names like "o1","o2","o3" and so on.
func (c *Collection) AddIndexed(prefix string, startIndex, endIndex int) error {
	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex)
	if err != nil {
		return err
	}
	ports, _ := indexedPorts.All()
	c.Add(ports...)
	return nil
}

// Signals returns all signals of all ports in the collection.
func (c *Collection) Signals() *signal.Group {
	group := signal.NewGroup()
	for _, p := range c.ports {
		signals, err := p.Signals().All()
		if err != nil {
			continue
		}
		group = group.With(signals...)
	}
	return group
}

// All returns all ports as a map.
func (c *Collection) All() (Map, error) {
	return c.ports, nil
}

// Any returns any arbitrary port from the collection.
// Returns nil if the collection is empty.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) Any() *Port {
	for _, port := range c.ports {
		return port
	}
	return nil
}

// Every returns true if all ports match the predicate.
func (c *Collection) Every(predicate Predicate) bool {
	for _, port := range c.ports {
		if !predicate(port) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any port matches the predicate.
// Note: AnyMatch is used instead of Any to avoid conflict with the no-arg Any() *Port method.
func (c *Collection) AnyMatch(predicate Predicate) bool {
	for _, port := range c.ports {
		if predicate(port) {
			return true
		}
	}
	return false
}

// Count returns the number of ports that match the predicate.
func (c *Collection) Count(predicate Predicate) int {
	count := 0
	for _, port := range c.ports {
		if predicate(port) {
			count++
		}
	}
	return count
}

// FindAny returns any arbitrary port that matches the predicate.
// Returns nil if no match found.
// Note: Map iteration order is not guaranteed, so this may return different items on each call.
func (c *Collection) FindAny(predicate Predicate) *Port {
	for _, port := range c.ports {
		if predicate(port) {
			return port
		}
	}
	return nil
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
	filtered := NewCollection()
	for _, port := range c.ports {
		if predicate(port) {
			filtered.Add(port)
		}
	}
	return filtered
}

// Map returns a new collection with ports transformed by the mapper function.
func (c *Collection) Map(mapper Mapper) *Collection {
	mapped := NewCollection()
	for _, port := range c.ports {
		if result := mapper(port); result != nil {
			mapped.Add(result)
		}
	}
	return mapped
}

// WithParentComponent sets the parent component on all ports in the collection and returns the collection.
func (c *Collection) WithParentComponent(comp ParentComponent) *Collection {
	for _, p := range c.ports {
		p.setParentComponent(comp)
	}
	return c
}

// Len returns the number of ports in a collection.
func (c *Collection) Len() int {
	return len(c.ports)
}

// IsEmpty returns true when there are no ports in the collection.
func (c *Collection) IsEmpty() bool {
	return c.Len() == 0
}
