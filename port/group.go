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
