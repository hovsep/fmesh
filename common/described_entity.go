package common

type DescribedEntity struct {
	description string
}

// NewDescribedEntity constructor
func NewDescribedEntity(description string) DescribedEntity {
	return DescribedEntity{description: description}
}

// Description getter
func (d DescribedEntity) Description() string {
	return d.description
}
