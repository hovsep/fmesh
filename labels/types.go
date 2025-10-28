package labels

// Map is a map of labels.
type Map map[string]string

// LabelPredicate tests a label key-value pair.
type LabelPredicate func(label, value string) bool

// LabelMapper transforms a label key-value pair into a new key-value pair.
type LabelMapper func(key, value string) (newKey, newValue string)
