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

// NewIndexedGroup is useful when you want to create group of ports with same prefix
func NewIndexedGroup(prefix string, startIndex int, endIndex int) Group {
	if prefix == "" {
		return nil
	}

	if startIndex > endIndex {
		return nil
	}

	group := make(Group, endIndex-startIndex+1)

	for i := startIndex; i <= endIndex; i++ {
		group[i-startIndex] = New(fmt.Sprintf("%s%d", prefix, i))
	}

	return group
}

// Add adds ports to group
func (group Group) Add(ports ...*Port) Group {
	for _, port := range ports {
		if port == nil {
			continue
		}
		group = append(group, port)
	}

	return group
}
