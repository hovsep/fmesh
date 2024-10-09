package common

type NamedEntity struct {
	name string
}

// NewNamedEntity constructor
func NewNamedEntity(name string) NamedEntity {
	return NamedEntity{name: name}
}

// Name getter
func (e NamedEntity) Name() string {
	return e.name
}
