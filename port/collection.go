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
	// Labels added by default to each port in a collection
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
	return c.AnyMatch(func(p *Port) bool {
		return p.HasSignals()
	})
}

// AllHaveSignals returns true when all ports in collection have signals.
func (c *Collection) AllHaveSignals() bool {
	return c.AllMatch(func(p *Port) bool {
		return p.HasSignals()
	})
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

	indexedPorts, err := NewIndexedGroup(prefix, startIndex, endIndex).AllAsSlice()
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
		signals, err := p.Signals().AllAsSlice()
		if err != nil {
			c.WithChainableErr(err)
			return signal.NewGroup().WithChainableErr(c.ChainableErr())
		}
		group = group.With(signals...)
	}
	return group
}

// AllAsMap returns all ports as a map.
func (c *Collection) AllAsMap() (Map, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	return c.ports, nil
}

// AllAsMapOrDefault returns all ports as map or the provided default.
func (c *Collection) AllAsMapOrDefault(defaultPorts Map) Map {
	ports, err := c.AllAsMap()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// AllAsMapOrNil returns all ports as map or nil in case of error.
func (c *Collection) AllAsMapOrNil() Map {
	return c.AllAsMapOrDefault(nil)
}

// AllAsSlice returns all ports as Ports wrapper type.
func (c *Collection) AllAsSlice() (Ports, error) {
	if c.HasChainableErr() {
		return nil, c.ChainableErr()
	}
	ports := make([]*Port, 0, len(c.ports))
	for _, port := range c.ports {
		ports = append(ports, port)
	}
	return Ports(ports), nil
}

// AllAsSliceOrDefault returns all ports as Ports wrapper or the provided default.
func (c *Collection) AllAsSliceOrDefault(defaultPorts Ports) Ports {
	ports, err := c.AllAsSlice()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// AllAsSliceOrNil returns all ports as Ports wrapper or nil in case of error.
func (c *Collection) AllAsSliceOrNil() Ports {
	return c.AllAsSliceOrDefault(nil)
}

// AllAsGroup returns all ports as a Group.
func (c *Collection) AllAsGroup() (*Group, error) {
	if c.HasChainableErr() {
		return NewGroup().WithChainableErr(c.ChainableErr()), c.ChainableErr()
	}
	ports := make([]*Port, 0, len(c.ports))
	for _, port := range c.ports {
		ports = append(ports, port)
	}
	return NewGroup().withPorts(ports), nil
}

// AllAsGroupOrDefault returns all ports as Group or the provided default.
func (c *Collection) AllAsGroupOrDefault(defaultGroup *Group) *Group {
	group, err := c.AllAsGroup()
	if err != nil {
		return defaultGroup
	}
	return group
}

// AllAsGroupOrNil returns all ports as Group or nil in case of error.
func (c *Collection) AllAsGroupOrNil() *Group {
	group, err := c.AllAsGroup()
	if err != nil {
		return nil
	}
	return group
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

// AnyOrDefault returns any arbitrary port or the provided default.
func (c *Collection) AnyOrDefault(defaultPort *Port) *Port {
	if c.HasChainableErr() || c.IsEmpty() {
		return defaultPort
	}
	for _, port := range c.ports {
		return port
	}
	return defaultPort
}

// AnyOrNil returns any arbitrary port or nil.
func (c *Collection) AnyOrNil() *Port {
	if c.HasChainableErr() || c.IsEmpty() {
		return nil
	}
	for _, port := range c.ports {
		return port
	}
	return nil
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

// NoneMatch returns true if no ports match the predicate.
func (c *Collection) NoneMatch(predicate Predicate) bool {
	return !c.AnyMatch(predicate)
}

// CountMatch returns the number of ports that match the predicate.
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

// Filter returns a new collection with ports that match the predicate.
func (c *Collection) Filter(predicate Predicate) *Collection {
	if c.HasChainableErr() {
		return NewCollection().WithChainableErr(c.ChainableErr())
	}
	filtered := NewCollection().WithDefaultLabels(c.defaultLabels)
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
	mapped := NewCollection().WithDefaultLabels(c.defaultLabels)
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

// ChainableErr returns chainable error.
func (c *Collection) ChainableErr() error {
	return c.chainableErr
}

// Len returns the number of ports in a collection.
func (c *Collection) Len() int {
	return len(c.ports)
}

// WithDefaultLabels adds default labels to all ports in collection.
func (c *Collection) WithDefaultLabels(labelMap labels.Map) *Collection {
	c.defaultLabels = labelMap
	return c
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
