package port

import (
	"fmt"

	"github.com/hovsep/fmesh/labels"
)

// Group represents a list of ports
// can carry multiple ports with the same name
// no lookup methods.
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
		ports[i] = New(name)
	}
	return newGroup.withPorts(ports)
}

// NewIndexedGroup is useful to create group of ports with same prefix
// NOTE: endIndex is inclusive, e.g. NewIndexedGroup("p", 0, 0) will create one port with name "p0".
func NewIndexedGroup(prefix string, startIndex, endIndex int) *Group {
	if startIndex > endIndex {
		return NewGroup().WithChainableErr(ErrInvalidRangeForIndexedGroup)
	}

	ports := make(Ports, endIndex-startIndex+1)

	for i := startIndex; i <= endIndex; i++ {
		ports[i-startIndex] = New(fmt.Sprintf("%s%d", prefix, i))
	}

	return NewGroup().withPorts(ports)
}

// With adds ports to group.
func (g *Group) With(ports ...*Port) *Group {
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

// withPorts sets ports.
func (g *Group) withPorts(ports Ports) *Group {
	g.ports = ports
	return g
}

// AllAsSlice returns all ports as Ports wrapper type.
func (g *Group) AllAsSlice() (Ports, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.ports, nil
}

// AllAsSliceOrNil returns ports as Ports wrapper or nil in case of any error.
func (g *Group) AllAsSliceOrNil() Ports {
	return g.AllAsSliceOrDefault(nil)
}

// AllAsSliceOrDefault returns ports as Ports wrapper or default in case of any error.
func (g *Group) AllAsSliceOrDefault(defaultPorts Ports) Ports {
	ports, err := g.AllAsSlice()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// AllAsCollection returns all ports as a Collection.
func (g *Group) AllAsCollection() (*Collection, error) {
	if g.HasChainableErr() {
		return NewCollection().WithChainableErr(g.ChainableErr()), g.ChainableErr()
	}
	collection := NewCollection()
	for _, port := range g.ports {
		collection = collection.With(port)
		if collection.HasChainableErr() {
			return collection, collection.ChainableErr()
		}
	}
	return collection, nil
}

// AllAsCollectionOrDefault returns all ports as Collection or the provided default.
func (g *Group) AllAsCollectionOrDefault(defaultCollection *Collection) *Collection {
	collection, err := g.AllAsCollection()
	if err != nil {
		return defaultCollection
	}
	return collection
}

// AllAsCollectionOrNil returns all ports as Collection or nil in case of error.
func (g *Group) AllAsCollectionOrNil() *Collection {
	collection, err := g.AllAsCollection()
	if err != nil {
		return nil
	}
	return collection
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

// ChainableErr returns chainable error.
func (g *Group) ChainableErr() error {
	return g.chainableErr
}

// Len returns the number of ports in a group.
func (g *Group) Len() int {
	return len(g.ports)
}

// WithPortLabels sets labels on each port within the group and returns it.
func (g *Group) WithPortLabels(labelMap labels.Map) *Group {
	for _, p := range g.AllAsSliceOrNil() {
		p.WithLabels(labelMap)
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
		return New("").WithChainableErr(g.ChainableErr())
	}
	if g.IsEmpty() {
		g.WithChainableErr(ErrNoPortsInGroup)
		return New("").WithChainableErr(g.ChainableErr())
	}
	return g.ports[0]
}

// FirstOrDefault returns the first port or the provided default.
func (g *Group) FirstOrDefault(defaultPort *Port) *Port {
	if g.HasChainableErr() || g.IsEmpty() {
		return defaultPort
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

// NoneMatch returns true if no ports match the predicate.
func (g *Group) NoneMatch(predicate Predicate) bool {
	return !g.AnyMatch(predicate)
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

// FirstMatch returns the first port that matches the predicate.
func (g *Group) FirstMatch(predicate Predicate) *Port {
	if g.HasChainableErr() {
		return New("").WithChainableErr(g.ChainableErr())
	}
	for _, port := range g.ports {
		if predicate(port) {
			return port
		}
	}
	g.WithChainableErr(ErrNoPortMatchesPredicate)
	return New("").WithChainableErr(g.ChainableErr())
}

// Filter returns a new group with ports that match the predicate.
func (g *Group) Filter(predicate Predicate) *Group {
	if g.HasChainableErr() {
		return NewGroup().WithChainableErr(g.ChainableErr())
	}
	filtered := NewGroup()
	for _, port := range g.ports {
		if predicate(port) {
			filtered = filtered.With(port)
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
			mapped = mapped.With(transformedPort)
			if mapped.HasChainableErr() {
				return mapped
			}
		}
	}
	return mapped
}
