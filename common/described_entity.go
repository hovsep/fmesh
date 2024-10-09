package common

type DescribedEntity struct {
	description string
}

// NewDescribedEntity constructor
func NewDescribedEntity(description string) DescribedEntity {
	return DescribedEntity{description: description}
}

// Description getter
func (e DescribedEntity) Description() string {
	return e.description
}
