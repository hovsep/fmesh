package port

// Group is just a slice of ports (useful to pass multiple ports as variadic argument)
type Group []*Port

// NewPortGroup creates multiple ports
func NewPortGroup(names ...string) Group {
	group := make(Group, len(names))
	for _, name := range names {
		group = append(group, NewPort(name))
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
