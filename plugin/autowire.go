package plugin

import (
	"errors"
	"fmt"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
)

// Autowire pipes components together by naming convention instead of by hand.
//
// A mesh of any size accumulates long stretches of wiring that say nothing an
// attentive reader could not have guessed: every component that keeps time gets
// the clock, every component that wants the weather gets the weather. Written
// out, that is a list which has to be maintained, and the failure mode is
// silent -- add a component, forget the list, and it sits there activating
// forever on inputs that never arrive, looking perfectly connected.
//
// The convention makes the list derivable. An input port named by Name(output)
// is wired to that output automatically, whatever it is and whenever it appears.
//
// Wiring happens as components are added, and in both directions: a new
// component is offered every output already in the mesh, and every output it
// brings is offered to the components already there. Order of AddComponents
// therefore does not matter, which is the property that makes it safe to lean on.
//
// What does matter is that a component carries its ports when it is added: the
// only moment a component is looked at is its arrival, so ports created later
// with AddInputs/AddOutputs are never wired, and -- as with any missing pipe --
// nothing says so. Declare ports in component.New.
//
// A mesh commonly wants more than one convention at once -- a clock that reaches
// everything keeping time, and a set of environmental factors that reach
// whatever asked for them by name. Those are separate rules, so they are
// separate plugins, which is why each carries its own PluginName.
type Autowire struct {
	// Name maps a source component and one of its output ports to the input
	// port name that should receive it. Returning "" declines to wire.
	Name func(source *component.Component, output *port.Port) string

	// PluginName distinguishes one convention from another on the same mesh.
	// Defaults to "autowire".
	PluginName string
}

// Prefixed wires an output to any input named "<prefix><component>_<port>".
//
// It is the shape most meshes converge on: "habitat_gas_environmental_gas" is
// the gas factor's environmental_gas output, and reads as one.
func Prefixed(prefix string) *Autowire {
	return &Autowire{
		PluginName: "autowire:prefixed:" + prefix,
		Name: func(source *component.Component, output *port.Port) string {
			return fmt.Sprintf("%s%s_%s", prefix, source.Name(), output.Name())
		},
	}
}

// Broadcast wires every output named portName to every input of the same name.
func Broadcast(portName string) *Autowire {
	return BroadcastAs(portName, portName)
}

// BroadcastAs wires every output named outputName to every input named
// inputName.
//
// This is the clock case, where the two differ: a component emitting "tick"
// feeds everything that declared an input called "time", including the
// components added after it.
func BroadcastAs(outputName, inputName string) *Autowire {
	return &Autowire{
		PluginName: "autowire:broadcast:" + outputName + "->" + inputName,
		Name: func(_ *component.Component, output *port.Port) string {
			if output.Name() != outputName {
				return ""
			}
			return inputName
		},
	}
}

// GetName implements fmesh.Plugin.
func (a *Autowire) GetName() string {
	if a.PluginName == "" {
		return "autowire"
	}
	return a.PluginName
}

// Init implements fmesh.Plugin.
func (a *Autowire) Init(fm *fmesh.FMesh) error {
	if a.Name == nil {
		return errors.New("autowire: Name must be set")
	}

	fm.SetupHooks(func(hooks *fmesh.Hooks) {
		hooks.OnComponentAdded(func(ctx *fmesh.ComponentAddedContext) error {
			arrived := ctx.Component

			return ctx.FMesh.Components().ForEach(func(existing *component.Component) error {
				if existing == arrived {
					// Wiring a component to itself would be a loopback, which is
					// never what a convention meant to express.
					return nil
				}
				if err := a.connect(existing, arrived); err != nil {
					return err
				}
				return a.connect(arrived, existing)
			})
		})
	})
	return nil
}

// connect pipes every output of source to the matching input of destination.
func (a *Autowire) connect(source, destination *component.Component) error {
	return source.Outputs().ForEach(func(out *port.Port) error {
		name := a.Name(source, out)
		if name == "" {
			return nil
		}
		in := destination.InputByName(name)
		if in == nil {
			return nil
		}
		if isPipedTo(out, in) {
			// Pipes are not deduplicated, and a port flushes once per pipe, so a
			// second identical pipe would deliver every signal twice. Two
			// conventions on the same mesh can easily agree on one pair.
			return nil
		}
		return out.PipeTo(in)
	})
}

// isPipedTo reports whether out already pipes to in.
func isPipedTo(out, in *port.Port) bool {
	return out.Pipes().Find(func(p *port.Port) bool { return p == in }) != nil
}
