package port

import "fmt"

// Group is just a slice of ports (useful to pass multiple ports as variadic argument)
type Group []*Port

// NewGroup creates multiple ports
func NewGroup(names ...string) Group {
	group := make(Group, len(names))
	for i, name := range names {
		group[i] = New(name)
	}
	return group
}

// NewIndexedGroup is useful to create group of ports with same prefix
// NOTE: endIndex is inclusive, e.g. NewIndexedGroup("p", 0, 0) will create one port with name "p0"
func NewIndexedGroup(prefix string, startIndex int, endIndex int) Group {
	if startIndex > endIndex {
		return nil
	}

	group := make(Group, endIndex-startIndex+1)

	for i := startIndex; i <= endIndex; i++ {
		group[i-startIndex] = New(fmt.Sprintf("%s%d", prefix, i))
	}

	return group
}

// With adds ports to group
func (group Group) With(ports ...*Port) Group {
	newGroup := make(Group, len(group)+len(ports))
	copy(newGroup, group)
	for i, port := range ports {
		newGroup[len(group)+i] = port
	}

	return newGroup
}
