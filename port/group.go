package port

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
