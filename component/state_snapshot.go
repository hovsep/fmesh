package component

import "github.com/hovsep/fmesh/port"

// StateSnapshot represents the state of a component (used by f-mesh to perform piping correctly)
type StateSnapshot struct {
	inputPortsMetadata  port.MetadataMap
	outputPortsMetadata port.MetadataMap
}

// NewStateSnapshot creates new component state snapshot
func NewStateSnapshot() *StateSnapshot {
	return &StateSnapshot{
		inputPortsMetadata:  make(port.MetadataMap),
		outputPortsMetadata: make(port.MetadataMap),
	}
}

// InputPortsMetadata ... getter
func (s *StateSnapshot) InputPortsMetadata() port.MetadataMap {
	return s.inputPortsMetadata
}

// OutputPortsMetadata ... getter
func (s *StateSnapshot) OutputPortsMetadata() port.MetadataMap {
	return s.outputPortsMetadata
}

// WithInputPortsMetadata sets important information about input ports
func (s *StateSnapshot) WithInputPortsMetadata(inputPortsMetadata port.MetadataMap) *StateSnapshot {
	s.inputPortsMetadata = inputPortsMetadata
	return s
}

// WithOutputPortsMetadata sets important information about output ports
func (s *StateSnapshot) WithOutputPortsMetadata(outputPortsMetadata port.MetadataMap) *StateSnapshot {
	s.outputPortsMetadata = outputPortsMetadata
	return s
}

// getStateSnapshot returns a snapshot of component state
func (c *Component) getStateSnapshot() *StateSnapshot {
	return NewStateSnapshot().
		WithInputPortsMetadata(c.Inputs().GetPortsMetadata()).
		WithOutputPortsMetadata(c.Outputs().GetPortsMetadata())
}
