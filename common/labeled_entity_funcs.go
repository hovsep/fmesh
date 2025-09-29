package common

// LabelPredicate tests a label key-value pair.
type LabelPredicate func(label, value string) bool

// All returns true if all labels in the collection satisfy the predicate.
func (e *LabeledEntity) All(labels LabelsCollection, pred LabelPredicate) bool {
	for k, v := range labels {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if any label in the collection satisfies the predicate.
func (e *LabeledEntity) Any(labels LabelsCollection, pred LabelPredicate) bool {
	for k, v := range labels {
		if pred(k, v) {
			return true
		}
	}
	return false
}
