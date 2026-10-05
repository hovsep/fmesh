package fmesh

import (
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
)

// Visitor receives the structure of a mesh from [FMesh.Walk]. Exporters
// implement it to render the mesh in their own format.
type Visitor interface {
	VisitMesh(fm *FMesh) error
	VisitComponent(c *component.Component) error
	VisitPort(c *component.Component, p *port.Port) error
	// VisitPipe is called once per pipe; both ports report their component
	// through ParentComponent.
	VisitPipe(from, to *port.Port) error
}

// Walk visits the mesh, then each component with its input and then output
// ports, then every pipe. Components and ports come in name order, and pipes by
// source port in that same order, each port's pipes in wiring order. So a walk is
// deterministic, and every port is visited before any pipe that touches it. Walk
// reads the mesh without changing it and stops at the first error.
func (fm *FMesh) Walk(v Visitor) error {
	if err := v.VisitMesh(fm); err != nil {
		return err
	}

	components := fm.Components().AllOrdered()
	for _, c := range components {
		if err := v.VisitComponent(c); err != nil {
			return err
		}
		for _, p := range append(c.Inputs().AllOrdered(), c.Outputs().AllOrdered()...) {
			if err := v.VisitPort(c, p); err != nil {
				return err
			}
		}
	}

	for _, c := range components {
		for _, out := range c.Outputs().AllOrdered() {
			if err := out.Pipes().ForEach(func(in *port.Port) error {
				return v.VisitPipe(out, in)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
