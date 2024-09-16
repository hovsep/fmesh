package cycle

// Collection contains the results of several activation cycles
type Collection []*Cycle

// NewCollection creates a collection
func NewCollection() Collection {
	return make(Collection, 0)
}

// Add adds cycle results to existing collection
func (collection Collection) Add(newCycleResults ...*Cycle) Collection {
	for _, cycleResult := range newCycleResults {
		collection = append(collection, cycleResult)
	}
	return collection
}
