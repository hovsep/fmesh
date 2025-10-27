package port

import (
	"fmt"

	"github.com/hovsep/fmesh/labels"
)

// Ports is a list of ports.
type Ports []*Port

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

// Ports getter.
func (g *Group) Ports() (Ports, error) {
	if g.HasChainableErr() {
		return nil, g.ChainableErr()
	}
	return g.ports, nil
}

// PortsOrNil returns ports or nil in case of any error.
func (g *Group) PortsOrNil() Ports {
	return g.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error.
func (g *Group) PortsOrDefault(defaultPorts Ports) Ports {
	ports, err := g.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
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
	for _, p := range g.PortsOrNil() {
		p.WithLabels(labelMap)
	}
	return g
}

// IsEmpty returns true when there are no ports in the group.
func (g *Group) IsEmpty() bool {
	return g.Len() == 0
}

// Any returns true if at least one signal matches the predicate.
func (g *Group) Any(p Predicate) bool {
	if g.HasChainableErr() {
		return false
	}

	if g.Len() == 0 {
		return false
	}

	for _, port := range g.PortsOrNil() {
		if p(port) {
			return true
		}
	}

	return false
}

// All returns true if all signals match the predicate.
func (g *Group) All(p Predicate) bool {
	if g.HasChainableErr() {
		return false
	}

	if g.IsEmpty() {
		return false
	}

	for _, port := range g.PortsOrNil() {
		if !p(port) {
			return false
		}
	}

	return true
}
