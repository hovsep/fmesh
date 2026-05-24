package meta

// Predicate tests a label key-value pair.
type Predicate func(label, value string) bool

// Mapper transforms a label key-value pair into a new key-value pair.
type Mapper func(key, value string) (newKey, newValue string)

// ScalarPredicate tests a scalar name-value pair.
type ScalarPredicate func(name string, value float64) bool
