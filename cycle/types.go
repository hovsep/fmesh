package cycle

// Predicate is a function that tests whether a Cycle matches a condition.
type Predicate func(cycle *Cycle) bool

// Mapper transforms a Cycle into a new Cycle.
type Mapper func(cycle *Cycle) *Cycle

// Cycles contain the results of several activation cycles.
type Cycles []*Cycle
