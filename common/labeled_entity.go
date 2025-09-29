package common

import (
	"errors"
	"fmt"
)

// LabelsCollection is a map of labels
type LabelsCollection map[string]string

// LabeledEntity is a base struct for labeled entities
type LabeledEntity struct {
	labels LabelsCollection
}

// ErrLabelNotFound is returned when a label is not found
var ErrLabelNotFound = errors.New("label not found")

// NewLabeledEntity constructor
func NewLabeledEntity(labels LabelsCollection) LabeledEntity {
	return LabeledEntity{labels: labels}
}

// Labels getter
func (e *LabeledEntity) Labels() LabelsCollection {
	return e.labels
}

// Label returns the value of a single label or empty string if it is not found
func (e *LabeledEntity) Label(label string) (string, error) {
	value, ok := e.labels[label]

	if !ok {
		return "", fmt.Errorf("label %s not found, %w", label, ErrLabelNotFound)
	}

	return value, nil
}

// LabelOrDefault returns label value or default value in case of any error
func (e *LabeledEntity) LabelOrDefault(label, defaultValue string) string {
	value, err := e.Label(label)
	if err != nil {
		return defaultValue
	}
	return value
}

// SetLabels overwrites a labels collection
func (e *LabeledEntity) SetLabels(labels LabelsCollection) {
	e.labels = labels
}

// AddLabel adds or updates(if label already exists) single label
func (e *LabeledEntity) AddLabel(label, value string) {
	if e.labels == nil {
		e.labels = make(LabelsCollection)
	}
	e.labels[label] = value
}

// AddLabels adds or updates(if label already exists) multiple labels
func (e *LabeledEntity) AddLabels(labels LabelsCollection) {
	for label, value := range labels {
		e.AddLabel(label, value)
	}
}

// DeleteLabels deletes given labels
func (e *LabeledEntity) DeleteLabels(labels ...string) {
	for _, label := range labels {
		delete(e.labels, label)
	}
}

// HasLabel returns true when the entity has given label or false otherwise
func (e *LabeledEntity) HasLabel(label string) bool {
	_, ok := e.labels[label]
	return ok
}

// HasAllLabels checks if an entity has all given labels.
func (e *LabeledEntity) HasAllLabels(label ...string) bool {
	labelsMap := make(LabelsCollection, len(label))
	for _, l := range label {
		labelsMap[l] = "" // value is ignored
	}
	return e.All(labelsMap, func(l, _ string) bool { return e.HasLabel(l) })
}

// HasAllLabelsWithValues returns true if the entity contains all key-value pairs from the given collection.
func (e *LabeledEntity) HasAllLabelsWithValues(labels LabelsCollection) bool {
	return e.All(labels, e.LabelIs)
}

// HasAnyLabel checks if an entity has any of the given labels.
func (e *LabeledEntity) HasAnyLabel(label ...string) bool {
	labelsMap := make(LabelsCollection, len(label))
	for _, l := range label {
		labelsMap[l] = ""
	}
	return e.Any(labelsMap, func(l, _ string) bool { return e.HasLabel(l) })
}

// HasAnyLabelWithValue returns true if the entity contains any key-value pair from the given collection.
func (e *LabeledEntity) HasAnyLabelWithValue(labels LabelsCollection) bool {
	return e.Any(labels, e.LabelIs)
}

// LabelIs returns true when an entity has given label with a given value
func (e *LabeledEntity) LabelIs(label, value string) bool {
	if !e.HasLabel(label) {
		return false
	}

	l, err := e.Label(label)
	if err != nil {
		return false
	}

	return l == value
}

// LabelsCount return the number of labels
func (e *LabeledEntity) LabelsCount() int {
	return len(e.Labels())
}
