package port

import (
	"context"

	"github.com/hovsep/fmesh/internal/collection"
	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/signal"
)

// keyedPorts hides the embedded field name so it cannot be reached or
// reassigned from outside; only the base's exported methods promote.
type keyedPorts = collection.Keyed[*Port]

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
	*keyedPorts
	labels  *meta.Labels
	scalars *meta.Scalars
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		keyedPorts: collection.NewKeyed[*Port]("port"),
		labels:     meta.NewLabels(),
		scalars:    meta.NewScalars(),
	}
}

// Labels returns the collection's own labels store.
func (c *Collection) Labels() *meta.Labels { return c.labels }

// Scalars returns the collection's own scalars store.
func (c *Collection) Scalars() *meta.Scalars { return c.scalars }

// ByNames retrieves a subset of ports by their names, returning a new collection.
// Names that match no port are skipped, which makes the result vacuously ready —
// see the note on forgiving name lookups in design.md.
func (c *Collection) ByNames(names ...string) *Collection {
	matched := make([]*Port, 0, len(names))
	for _, name := range names {
		if p := c.ByName(name); p != nil {
			matched = append(matched, p)
		}
	}
	selected := NewCollection()
	_ = selected.Add(matched...) // ports come from existing collection — names are unique by construction
	return selected
}

// AnyHasSignals returns true if at least one port in collection has signals.
func (c *Collection) AnyHasSignals() bool {
	return c.AnyMatch(func(p *Port) bool {
		return p.HasSignals()
	})
}

// AllHaveSignals returns true when all ports in the collection have signals.
func (c *Collection) AllHaveSignals() bool {
	return c.Every(func(p *Port) bool {
		return p.HasSignals()
	})
}

// PutSignalsOnEach adds the same signals to every port in the collection.
// The name says OnEach because this broadcasts — it is not a fan-in.
// Stops and returns the first error encountered.
func (c *Collection) PutSignalsOnEach(signals ...*signal.Signal) error {
	for p := range c.Each {
		if err := p.PutSignals(signals...); err != nil {
			return err
		}
	}
	return nil
}

// Flush flushes all ports in a collection.
// Stops and returns the first error encountered.
func (c *Collection) Flush(ctx context.Context) error {
	for p := range c.Each {
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
	for p := range c.Each {
		if err := p.PipeTo(destPorts...); err != nil {
			return err
		}
	}
	return nil
}

// Remove deletes ports by name and returns the collection.
func (c *Collection) Remove(names ...string) *Collection {
	c.keyedPorts.Remove(names...)
	return c
}

// Signals returns all signals of all ports in the collection.
func (c *Collection) Signals() *signal.Group {
	group := signal.NewGroup()
	for p := range c.Each {
		signals := p.Signals().All()
		group = group.With(signals...)
	}
	return group
}

// Any returns the first port in the collection by name order.
// Returns nil if the collection is empty.
// The name says "any" because callers should not depend on which one they get;
// it is nonetheless stable across runs, like every traversal here.
func (c *Collection) Any() *Port {
	for p := range c.Each {
		return p
	}
	return nil
}

// Filter returns a new collection containing only ports that match the predicate.
func (c *Collection) Filter(predicate Predicate) *Collection {
	matched := make([]*Port, 0, c.Len())
	for port := range c.Each {
		if predicate(port) {
			matched = append(matched, port)
		}
	}
	filtered := NewCollection()
	_ = filtered.Add(matched...) // ports come from existing collection — names are unique by construction
	return filtered
}

// Map returns a new collection with ports transformed by the mapper function.
// Returns an error if a mapped port has a duplicate name.
func (c *Collection) Map(mapper Mapper) (*Collection, error) {
	transformed := make([]*Port, 0, c.Len())
	for port := range c.Each {
		if transformedPort := mapper(port); transformedPort != nil {
			transformed = append(transformed, transformedPort)
		}
	}
	mapped := NewCollection()
	if err := mapped.Add(transformed...); err != nil {
		return nil, err
	}
	return mapped, nil
}

// SetParentComponent sets the parent component on all ports in the collection and returns the collection.
func (c *Collection) SetParentComponent(comp ParentComponent) *Collection {
	for p := range c.Each {
		p.setParentComponent(comp)
	}
	return c
}
