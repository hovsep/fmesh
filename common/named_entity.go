package common

type NamedEntity struct {
	name string
}

// NewNamedEntity constructor
func NewNamedEntity(name string) NamedEntity {
	return NamedEntity{name: name}
}

// Name getter
func (n NamedEntity) Name() string {
	return n.name
}
