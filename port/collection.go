package port

import (
	"github.com/hovsep/fmesh/internal/collection"
	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/signal"
)

// keyedPorts hides the embedded field name so it cannot be reached or
// reassigned from outside; only the base's exported methods promote.
type keyedPorts = collection.Keyed[*Port]

// Collection is a port collection indexed by name; it cannot carry two ports
// with the same name. Optimized for lookups.
//
// Every traversal goes in port-name order — map order would leak into flush and
// Signals() results and break run determinism.
type Collection struct {
	*keyedPorts
	meta *meta.Meta
}

// NewCollection creates an empty collection.
func NewCollection() *Collection {
	return &Collection{
		keyedPorts: collection.NewKeyed[*Port]("port"),
		meta:       meta.New(),
	}
}

// Meta returns the collection's own metadata store.
func (c *Collection) Meta() *meta.Meta { return c.meta }

// ByNames retrieves a subset of ports by their names, returning a new collection.
// Names that match no port are skipped, which makes the result vacuously ready.
// Duplicated names collapse to one lookup, because the batched Add rejects a repeated name and would add none of
// them.
func (c *Collection) ByNames(names ...string) *Collection {
	matched := make([]*Port, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		if p := c.ByName(name); p != nil {
			matched = append(matched, p)
		}
	}
	selected := NewCollection()
	_ = selected.Add(matched...) // deduped above, taken from one collection — no conflict possible
	return selected
}

// AnyHasSignals returns true if at least one port in collection has signals.
func (c *Collection) AnyHasSignals() bool {
	return c.Any(func(p *Port) bool {
		return p.HasSignals()
	})
}

// AllHaveSignals returns true when all ports in the collection have signals.
func (c *Collection) AllHaveSignals() bool {
	return c.Every(func(p *Port) bool {
		return p.HasSignals()
	})
}

// Remove deletes ports by name and returns the collection.
func (c *Collection) Remove(names ...string) *Collection {
	c.keyedPorts.Remove(names...)
	return c
}

// Signals returns all signals of all ports in the collection.
func (c *Collection) Signals() *signal.Group {
	signals := make([]*signal.Signal, 0, c.Len())
	for p := range c.Each {
		signals = append(signals, p.Signals().All()...)
	}
	return signal.NewGroup().With(signals...)
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

// SetParentComponent sets the parent component on all ports in the collection and returns the collection.
func (c *Collection) SetParentComponent(comp ParentComponent) *Collection {
	for p := range c.Each {
		p.setParentComponent(comp)
	}
	return c
}
