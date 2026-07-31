package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
)

// Autowire pipes components together by naming convention instead of by hand: an
// input port named by Name(output) is wired to that output. See the wiki page on
// plugins for when to reach for it.
//
// Wiring happens in both directions as components are added, so the order of
// AddComponents does not matter. A component must carry its ports when it is
// added, though — arrival is the only moment it is looked at, so ports added
// later with AddInputs/AddOutputs are never wired, and nothing reports it.
//
// One convention per instance, each with its own PluginName; a mesh that wants
// two rules registers two plugins.
type Autowire struct {
	// Name maps a source component and one of its output ports to the input
	// port name that should receive it. Returning "" declines to wire.
	Name func(source *component.Component, output *port.Port) string

	// PluginName distinguishes one convention from another on the same mesh.
	// Defaults to "autowire".
	PluginName string
}

// AutowirePrefixed wires an output to any input named "<prefix><component>_<port>".
//
// It is the shape most meshes converge on: "habitat_gas_environmental_gas" is
// the gas factor's environmental_gas output, and reads as one.
func AutowirePrefixed(prefix string) *Autowire {
	return &Autowire{
		PluginName: "autowire:prefixed:" + prefix,
		Name: func(source *component.Component, output *port.Port) string {
			return fmt.Sprintf("%s%s_%s", prefix, source.Name(), output.Name())
		},
	}
}

// AutowireBroadcast wires every output named portName to every input of the
// same name.
func AutowireBroadcast(portName string) *Autowire {
	return AutowireBroadcastAs(portName, portName)
}

// AutowireBroadcastAs wires every output named outputName to every input named
// inputName.
//
// This is the clock case, where the two differ: a component emitting "tick"
// feeds everything that declared an input called "time", including the
// components added after it.
func AutowireBroadcastAs(outputName, inputName string) *Autowire {
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
		hooks.OnComponentAdded(func(_ context.Context, added *fmesh.ComponentAddedContext) error {
			arrived := added.Component

			return added.FMesh.Components().ForEach(func(existing *component.Component) error {
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
