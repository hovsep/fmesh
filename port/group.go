package port

import (
	"fmt"
	"github.com/hovsep/fmesh/common"
)

// Group represents a list of ports
// can carry multiple ports with same name
// no lookup methods
type Group struct {
	*common.Chainable
	ports []*Port //@TODO: extract type
}

// NewGroup creates multiple ports
func NewGroup(names ...string) *Group {
	newGroup := &Group{
		Chainable: common.NewChainable(),
	}
	ports := make([]*Port, len(names))
	for i, name := range names {
		ports[i] = New(name)
	}
	return newGroup.withPorts(ports)
}

// NewIndexedGroup is useful to create group of ports with same prefix
// NOTE: endIndex is inclusive, e.g. NewIndexedGroup("p", 0, 0) will create one port with name "p0"
func NewIndexedGroup(prefix string, startIndex int, endIndex int) *Group {
	if startIndex > endIndex {
		return nil
	}

	ports := make([]*Port, endIndex-startIndex+1)

	for i := startIndex; i <= endIndex; i++ {
		ports[i-startIndex] = New(fmt.Sprintf("%s%d", prefix, i))
	}

	return NewGroup().withPorts(ports)
}

// With adds ports to group
func (g *Group) With(ports ...*Port) *Group {
	if g.HasChainError() {
		return g
	}

	newPorts := make([]*Port, len(g.ports)+len(ports))
	copy(newPorts, g.ports)
	for i, port := range ports {
		newPorts[len(g.ports)+i] = port
	}

	return g.withPorts(newPorts)
}

// withPorts sets ports
func (g *Group) withPorts(ports []*Port) *Group {
	g.ports = ports
	return g
}

// Ports getter
func (g *Group) Ports() ([]*Port, error) {
	if g.HasChainError() {
		return nil, g.ChainError()
	}
	return g.ports, nil
}

// PortsOrNil returns ports or nil in case of any error
func (g *Group) PortsOrNil() []*Port {
	return g.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error
func (g *Group) PortsOrDefault(defaultPorts []*Port) []*Port {
	ports, err := g.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// WithChainError returns group with error
func (g *Group) WithChainError(err error) *Group {
	g.SetChainError(err)
	return g
}

// Len returns number of ports in group
func (g *Group) Len() int {
	return len(g.ports)
}
