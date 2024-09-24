package cycle

// Collection contains the results of several activation cycles
type Collection []*Cycle

// NewCollection creates a collection
func NewCollection() Collection {
	return make(Collection, 0)
}

// With adds cycle results to existing collection
func (collection Collection) With(cycleResults ...*Cycle) Collection {
	newCollection := make(Collection, len(collection)+len(cycleResults))
	copy(newCollection, collection)
	for i, cycleResult := range cycleResults {
		newCollection[len(collection)+i] = cycleResult
	}
	return newCollection
}
