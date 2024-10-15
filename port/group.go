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
	ports []*Port
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
func (group *Group) With(ports ...*Port) *Group {
	if group.HasError() {
		return group
	}

	newPorts := make([]*Port, len(group.ports)+len(ports))
	copy(newPorts, group.ports)
	for i, port := range ports {
		newPorts[len(group.ports)+i] = port
	}

	return group.withPorts(newPorts)
}

// withPorts sets ports
func (group *Group) withPorts(ports []*Port) *Group {
	group.ports = ports
	return group
}

// Ports getter
func (group *Group) Ports() ([]*Port, error) {
	if group.HasError() {
		return nil, group.Error()
	}
	return group.ports, nil
}

// PortsOrNil returns ports or nil in case of any error
func (group *Group) PortsOrNil() []*Port {
	return group.PortsOrDefault(nil)
}

// PortsOrDefault returns ports or default in case of any error
func (group *Group) PortsOrDefault(defaultPorts []*Port) []*Port {
	ports, err := group.Ports()
	if err != nil {
		return defaultPorts
	}
	return ports
}

// WithError returns group with error
func (group *Group) WithError(err error) *Group {
	group.SetError(err)
	return group
}

// Len returns number of ports in group
func (group *Group) Len() int {
	return len(group.ports)
}
