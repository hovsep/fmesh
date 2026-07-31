package port

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/signal"
)

// Collection is a port collection.
// indexed by name; hence it cannot carry
// 2 ports with the same name. Optimized for lookups.
//
// Every traversal goes in port-name order. Map iteration order would otherwise
// leak into results: the order in which a component's output ports are flushed
// decides the order their signals arrive at a shared downstream port, and
// Signals() decides what a component reading all its inputs at once sees. Both
// were observably random before — the same mesh, run twice, produced different
// answers. See .agent/docs/design.md for the ordering guarantees this upholds.
type Collection struct {
	// ports indexes by name, for ByName.
	ports map[string]*Port
	// ordered holds the same ports, sorted by name — traversal reads this, so it
	// costs neither a per-element map lookup nor an allocation. It is rebuilt when
	// membership changes rather than lazily on read: ports are read from
	// activation goroutines, and a lazily filled cache would turn a read into a
	// write.
	ordered []*Port
	labels  *meta.Labels
	scalars *meta.Scalars
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		ports:   make(map[string]*Port),
		labels:  meta.NewLabels(),
		scalars: meta.NewScalars(),
	}
}

// rebuildOrder refreshes the sorted port list. Called on every membership change,
// which in practice means construction.
func (c *Collection) rebuildOrder() {
	c.ordered = make([]*Port, 0, len(c.ports))
	for _, name := range slices.Sorted(maps.Keys(c.ports)) {
		c.ordered = append(c.ordered, c.ports[name])
	}
}

// AllOrdered returns all ports sorted by name.
// This is the order every traversal on a Collection uses.
// A copy is returned so the caller cannot reorder the collection's own state;
// inside the package range over each instead.
func (c *Collection) AllOrdered() []*Port {
	return slices.Clone(c.ordered)
}

// each yields every port in name order, without allocating and without a map
// lookup per port.
//
// This is the internal form of AllOrdered, and the one the traversals below use:
// ClearInputs and FlushOutputs walk a collection for every component on every
// cycle, so neither a slice copy nor a hash lookup per port belongs there.
// Range over it: for p := range c.each.
func (c *Collection) each(yield func(*Port) bool) {
	for _, p := range c.ordered {
		if !yield(p) {
			return
		}
	}
}

// Labels returns the collection's own labels store.
func (c *Collection) Labels() *meta.Labels { return c.labels }

// Scalars returns the collection's own scalars store.
func (c *Collection) Scalars() *meta.Scalars { return c.scalars }

// SetLabelOnEach sets a label on every port in the collection.
func (c *Collection) SetLabelOnEach(name, value string) *Collection {
	for p := range c.each {
		p.labels.Set(name, value)
	}
	return c
}

// SetScalarOnEach sets a scalar on every port in the collection.
func (c *Collection) SetScalarOnEach(name string, value float64) *Collection {
	for p := range c.each {
		p.scalars.Set(name, value)
	}
	return c
}

// RemoveLabelOnEach removes a label from every port in the collection.
func (c *Collection) RemoveLabelOnEach(names ...string) *Collection {
	for p := range c.each {
		p.labels.Remove(names...)
	}
	return c
}

// RemoveScalarOnEach removes a scalar from every port in the collection.
func (c *Collection) RemoveScalarOnEach(names ...string) *Collection {
	for p := range c.each {
		p.scalars.Remove(names...)
	}
	return c
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
			_ = selected.Add(p) // port comes from existing collection — name is unique by construction
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

// PutSignalsOnEach adds the same signals to every port in the collection.
// The name says OnEach because this broadcasts — it is not a fan-in.
// Stops and returns the first error encountered.
func (c *Collection) PutSignalsOnEach(signals ...*signal.Signal) error {
	for p := range c.each {
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
	for p := range c.each {
		if err := action(p); err != nil {
			return err
		}
	}
	return nil
}

// Flush flushes all ports in a collection.
// Stops and returns the first error encountered.
func (c *Collection) Flush(ctx context.Context) error {
	for p := range c.each {
		if err := p.Flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

// PipeEachTo pipes every port in the collection to every destination port —
// a full cross product, which the name makes explicit.
// Stops and returns the first error encountered.
func (c *Collection) PipeEachTo(destPorts ...*Port) error {
	for p := range c.each {
		if err := p.PipeTo(destPorts...); err != nil {
			return err
		}
	}
	return nil
}

// Add adds ports to the collection. Returns an error on name conflict.
func (c *Collection) Add(ports ...*Port) error {
	for _, p := range ports {
		if _, exists := c.ports[p.Name()]; exists {
			return fmt.Errorf("port %q already exists in collection", p.Name())
		}
		c.ports[p.Name()] = p
	}
	c.rebuildOrder()
	return nil
}

// Remove deletes ports by name and returns the collection.
func (c *Collection) Remove(names ...string) *Collection {
	for _, name := range names {
		delete(c.ports, name)
	}
	c.rebuildOrder()
	return c
}

// Signals returns all signals of all ports in the collection.
func (c *Collection) Signals() *signal.Group {
	group := signal.NewGroup()
	for p := range c.each {
		signals := p.Signals().All()
		group = group.With(signals...)
	}
	return group
}

// All returns a shallow copy of all ports as a map.
// A copy is returned so the caller cannot mutate the internal state.
func (c *Collection) All() map[string]*Port {
	return maps.Clone(c.ports)
}

// Any returns the first port in the collection by name order.
// Returns nil if the collection is empty.
// The name says "any" because callers should not depend on which one they get;
// it is nonetheless stable across runs, like every traversal here.
func (c *Collection) Any() *Port {
	for port := range c.each {
		return port
	}
	return nil
}

// Every returns true if all ports match the predicate.
func (c *Collection) Every(predicate Predicate) bool {
	for port := range c.each {
		if !predicate(port) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any port matches the predicate.
// Note: AnyMatch is used instead of Any to avoid conflict with the no-arg Any() *Port method.
func (c *Collection) AnyMatch(predicate Predicate) bool {
	for port := range c.each {
		if predicate(port) {
			return true
		}
	}
	return false
}

// Count returns the number of ports that match the predicate.
func (c *Collection) Count(predicate Predicate) int {
	count := 0
	for port := range c.each {
		if predicate(port) {
			count++
		}
	}
	return count
}

// FindAny returns the first port matching the predicate, in name order.
// Returns nil if no match found.
func (c *Collection) FindAny(predicate Predicate) *Port {
	for port := range c.each {
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
//	    labels := p.Labels().All()
//	    return labels["priority"] == "high"
//	})
func (c *Collection) Filter(predicate Predicate) *Collection {
	filtered := NewCollection()
	for port := range c.each {
		if predicate(port) {
			_ = filtered.Add(port) // port comes from existing collection — name is unique by construction
		}
	}
	return filtered
}

// Map returns a new collection with ports transformed by the mapper function.
// Returns an error if a mapped port has a duplicate name.
func (c *Collection) Map(mapper Mapper) (*Collection, error) {
	mapped := NewCollection()
	for port := range c.each {
		transformedPort := mapper(port)
		if transformedPort != nil {
			if err := mapped.Add(transformedPort); err != nil {
				return nil, err
			}
		}
	}
	return mapped, nil
}

// SetParentComponent sets the parent component on all ports in the collection and returns the collection.
func (c *Collection) SetParentComponent(comp ParentComponent) *Collection {
	for p := range c.each {
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
