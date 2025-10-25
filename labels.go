package fmesh

import (
	"errors"
	"fmt"

	"github.com/hovsep/fmesh/common"
)

// LabelsMap is a map of labels.
type LabelsMap map[string]string

type LabelsCollection struct {
	*common.Chainable
	labels LabelsMap
}

// LabelPredicate tests a label key-value pair.
type LabelPredicate func(label, value string) bool

// ErrLabelNotFound is returned when a label is not found.
var ErrLabelNotFound = errors.New("label not found")

// All returns true if all labels in the collection satisfy the predicate.
func (lc *LabelsCollection) All(labels LabelsMap, pred LabelPredicate) bool {
	for k, v := range labels {
		if !pred(k, v) {
			return false
		}
	}
	return true
}

// Any returns true if any label in the collection satisfies the predicate.
func (lc *LabelsCollection) Any(labels LabelsMap, pred LabelPredicate) bool {
	for k, v := range labels {
		if pred(k, v) {
			return true
		}
	}
	return false
}

// Label returns the value of a single label or empty string if it is not found.
func (lc *LabelsCollection) Label(label string) (string, error) {
	value, ok := lc.labels[label]

	if !ok {
		return "", fmt.Errorf("label %s not found, %w", label, ErrLabelNotFound)
	}

	return value, nil
}

// LabelOrDefault returns label value or default value in case of any error.
func (lc *LabelsCollection) LabelOrDefault(label, defaultValue string) string {
	value, err := lc.Label(label)
	if err != nil {
		return defaultValue
	}
	return value
}

// SetLabels overwrites a labels map.
func (lc *LabelsCollection) SetLabels(labels LabelsMap) {
	lc.labels = labels
}

// AddLabel adds or updates(if label already exists) single label.
func (lc *LabelsCollection) AddLabel(label, value string) {
	if lc.labels == nil {
		lc.labels = make(LabelsMap)
	}
	lc.labels[label] = value
}

// AddLabels adds or updates(if label already exists) multiple labels.
func (lc *LabelsCollection) AddLabels(labels LabelsMap) {
	for label, value := range labels {
		lc.AddLabel(label, value)
	}
}

// DeleteLabels deletes given labels.
func (lc *LabelsCollection) DeleteLabels(labels ...string) {
	for _, label := range labels {
		delete(lc.labels, label)
	}
}

// HasLabel returns true when the entity has given label or false otherwise.
func (lc *LabelsCollection) HasLabel(label string) bool {
	_, ok := lc.labels[label]
	return ok
}

// HasAllLabels checks if an entity has all given labels.
func (lc *LabelsCollection) HasAllLabels(label ...string) bool {
	labelsMap := make(LabelsMap, len(label))
	for _, l := range label {
		labelsMap[l] = "" // value is ignored
	}
	return lc.All(labelsMap, func(l, _ string) bool { return lc.HasLabel(l) })
}

// HasAllLabelsWithValues returns true if the entity contains all key-value pairs from the given collection.
func (lc *LabelsCollection) HasAllLabelsWithValues(labels LabelsMap) bool {
	return lc.All(labels, lc.LabelIs)
}

// HasAnyLabel checks if an entity has any of the given labels.
func (lc *LabelsCollection) HasAnyLabel(label ...string) bool {
	labelsMap := make(LabelsMap, len(label))
	for _, l := range label {
		labelsMap[l] = ""
	}
	return lc.Any(labelsMap, func(l, _ string) bool { return lc.HasLabel(l) })
}

// HasAnyLabelWithValue returns true if the entity contains any key-value pair from the given collection.
func (lc *LabelsCollection) HasAnyLabelWithValue(labels LabelsMap) bool {
	return lc.Any(labels, lc.LabelIs)
}

// LabelIs returns true when an entity has given label with a given value.
func (lc *LabelsCollection) LabelIs(label, value string) bool {
	if !lc.HasLabel(label) {
		return false
	}

	l, err := lc.Label(label)
	if err != nil {
		return false
	}

	return l == value
}

// LabelsCount return the number of labels.
func (lc *LabelsCollection) LabelsCount() int {
	return len(lc.labels)
}
