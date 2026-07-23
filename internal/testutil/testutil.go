// Package testutil provides panic-on-error constructors shared by the
// integration test suites. Not usable from in-package unit tests of fmesh or
// port (import cycle); those keep small local equivalents.
package testutil

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

// MustComponent builds a component or panics.
func MustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// MustFMesh builds a mesh or panics.
func MustFMesh(name string, opts ...fmesh.Option) *fmesh.FMesh {
	fm, err := fmesh.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return fm
}

// MustInputPort builds an input port or panics.
func MustInputPort(name string, opts ...port.Option) *port.Port {
	p, err := port.NewInput(name, opts...)
	if err != nil {
		panic(err)
	}
	return p
}

// MustOutputPort builds an output port or panics.
func MustOutputPort(name string, opts ...port.Option) *port.Port {
	p, err := port.NewOutput(name, opts...)
	if err != nil {
		panic(err)
	}
	return p
}

// MustPutSignals puts signals on a port or panics.
func MustPutSignals(p *port.Port, signals ...*signal.Signal) {
	if err := p.PutSignals(signals...); err != nil {
		panic(err)
	}
}

// MustPipeTo pipes src to dsts or panics.
func MustPipeTo(src *port.Port, dsts ...*port.Port) {
	if err := src.PipeTo(dsts...); err != nil {
		panic(err)
	}
}
