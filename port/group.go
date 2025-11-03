package port

import (
	"fmt"

	"github.com/hovsep/fmesh/labels"
)

// Group represents a list of ports.
// It can carry multiple ports with the same name and has no lookup methods.
type Group struct {
	chainableErr error
	ports        Ports
}

// NewGroup creates multiple ports.
func NewGroup(names ...string) *Group {
	newGroup := &Group{
		chainableErr: nil,
	}
	ports := make(Ports, len(names))
	for i, name := range names {
		ports[i] = NewOutput(name)
	}
	return newGroup.withPorts(ports)
}

// NewIndexedGroup creates a group of ports with the same prefix.
// NOTE: endIndex is inclusive, e.g. NewIndexedGroup("p", 0, 0) will create one port with name "p0".
func NewIndexedGroup(prefix string, startIndex, endIndex int) *Group {
	if startIndex > endIndex {
		return NewGroup().WithChainableErr(ErrInvalidRangeForIndexedGroup)
	}

	ports := make(Ports, endIndex-startIndex+1)

	for i := startIndex; i <= endIndex; i++ {
		ports[i-startIndex] = NewOutput(fmt.Sprintf("%s%d", prefix, i))
	}

	return NewGroup().withPorts(ports)
}

// Add adds ports to group.
func (g *Group) Add(ports ...*Port) *Group {
	if g.HasChainableErr() {
		return g
	}

	newPorts := make(Ports, len(g.ports)+len(ports))
	copy(newPorts, g.ports)
	for i, port := range ports {
		newPorts[len(g.ports)+i] = port
	}

	return g.withPorts(newPorts)
}

// Without removes ports matching the predicate and returns a new group.
func (g *Group) Without(predicate Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	// Keep ports that DON'T match the predicate
	return g.Filter(func(p *Port) bool {
		return !predicate(p)
	})
}

// ForEach applies the action to each port and returns the group for chaining.
func (g *Group) ForEach(action func(*Port)) *Group {
	if g.HasChainableErr() {
		return g
	}
	for _, p := range g.ports {
		action(p)
	}
	return g
}

// withPorts sets ports.
func (g *Group) withPorts(ports Ports) *Group {
	g.ports = ports
	return g
}

// All returns all ports as a slice.
func (g *Group) All() (Ports, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.ports, nil
}

// WithChainableErr sets a chainable error and returns the group.
func (g *Group) WithChainableErr(err error) *Group {
	g.chainableErr = err
	return g
}

// HasChainableErr returns true when a chainable error is set.
func (g *Group) HasChainableErr() bool {
	return g.chainableErr != nil
}

// ChainableErr returns the chainable error.
func (g *Group) ChainableErr() error {
	return g.chainableErr
}

// Len returns the number of ports in a group.
func (g *Group) Len() int {
	return len(g.ports)
}

// AddLabelsToAll adds labels to each port within the group and returns it.
func (g *Group) AddLabelsToAll(labelMap labels.Map) *Group {
	for _, p := range g.ports {
		p.AddLabels(labelMap)
	}
	return g
}

// IsEmpty returns true when there are no ports in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// First returns the first port in the group.
func (g *Group) First() *Port {
	if g.HasChainableErr() {
		return NewOutput("").WithChainableErr(g.ChainableErr())
	}
	if g.IsEmpty() {
		g.WithChainableErr(ErrNoPortsInGroup)
		return NewOutput("").WithChainableErr(g.ChainableErr())
	}
	return g.ports[0]
}

// FirstOrNil returns the first port or nil.
func (g *Group) FirstOrNil() *Port {
	if g.HasChainableErr() || g.IsEmpty() {
		return nil
	}
	return g.ports[0]
}

// AllMatch returns true if all ports match the predicate.
func (g *Group) AllMatch(predicate Predicate) bool {
	if g.HasChainableErr() {
		return false
	}
	for _, port := range g.ports {
		if !predicate(port) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any port matches the predicate.
func (g *Group) AnyMatch(predicate Predicate) bool {
	if g.HasChainableErr() {
		return false
	}
	for _, port := range g.ports {
		if predicate(port) {
			return true
		}
	}
	return false
}

// CountMatch returns the number of ports that match the predicate.
func (g *Group) CountMatch(predicate Predicate) int {
	if g.HasChainableErr() {
		return 0
	}
	count := 0
	for _, port := range g.ports {
		if predicate(port) {
			count++
		}
	}
	return count
}

// Filter returns a new group with ports that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	filtered := NewGroup()
	for _, port := range g.ports {
		if predicate(port) {
			filtered = filtered.Add(port)
			if filtered.HasChainableErr() {
				return filtered
			}
		}
	}
	return filtered
}

// Map returns a new group with ports transformed by the mapper function.
func (g *Group) Map(mapper Mapper) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	mapped := NewGroup()
	for _, port := range g.ports {
		transformedPort := mapper(port)
		if transformedPort != nil {
			mapped = mapped.Add(transformedPort)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
}
