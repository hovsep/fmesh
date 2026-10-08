package cycle

// Predicate is a function that tests whether a Cycle matches a condition.
type Predicate func(cycle *Cycle) bool
