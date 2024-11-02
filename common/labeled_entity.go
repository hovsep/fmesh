package common

import (
	"errors"
	"fmt"
)

type LabelsCollection map[string]string

type LabeledEntity struct {
	labels LabelsCollection
}

var (
	ErrLabelNotFound = errors.New("label not found")
)

// NewLabeledEntity constructor
func NewLabeledEntity(labels LabelsCollection) LabeledEntity {
	return LabeledEntity{labels: labels}
}

// Labels getter
func (e *LabeledEntity) Labels() LabelsCollection {
	return e.labels
}

// Label returns the value of single label or nil if it is not found
func (e *LabeledEntity) Label(label string) (string, error) {
	value, ok := e.labels[label]

	if !ok {
		return "", fmt.Errorf("label %s not found, %w", label, ErrLabelNotFound)
	}

	return value, nil
}

// LabelOrDefault returns label value or default value in case of any error
func (e *LabeledEntity) LabelOrDefault(label string, defaultValue string) string {
	value, err := e.Label(label)
	if err != nil {
		return defaultValue
	}
	return value
}

// SetLabels overwrites labels collection
func (e *LabeledEntity) SetLabels(labels LabelsCollection) {
	e.labels = labels
}

// AddLabel adds or updates(if label already exists) single label
func (e *LabeledEntity) AddLabel(label string, value string) {
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

// DeleteLabel deletes given label
func (e *LabeledEntity) DeleteLabel(label string) {
	delete(e.labels, label)
}

// HasLabel returns true when entity has given label or false otherwise
func (e *LabeledEntity) HasLabel(label string) bool {
	_, ok := e.labels[label]
	return ok
}

// HasAllLabels checks if entity has all labels
func (e *LabeledEntity) HasAllLabels(label ...string) bool {
	for _, l := range label {
		if !e.HasLabel(l) {
			return false
		}
	}
	return true
}

// HasAnyLabel checks if entity has at least one of given labels
func (e *LabeledEntity) HasAnyLabel(label ...string) bool {
	for _, l := range label {
		if e.HasLabel(l) {
			return true
		}
	}
	return false
}
