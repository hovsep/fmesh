package autowire

import (
	"context"
	"errors"
	"fmt"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
)

// Plugin pipes components together by naming convention instead of by hand: an
// input port named by InputNameFor(source, output) is wired to that output. See
// the wiki page on plugins for when to reach for it.
//
// Wiring happens in both directions as components are added, so the order of
// AddComponents does not matter. A component must carry its ports when it is
// added, though — arrival is the only moment it is looked at, so ports added
// later with AddInputs/AddOutputs are never wired, and nothing reports it.
//
// One convention per instance, each with its own PluginName; a mesh that wants
// two rules registers two plugins.
type Plugin struct {
	// InputNameFor maps a source component and one of its output ports to the
	// input port name that should receive it. Returning "" declines to wire.
	InputNameFor func(source *component.Component, output *port.Port) string

	// PluginName distinguishes one convention from another on the same mesh.
	// Defaults to "autowire".
	PluginName string
}

// Prefixed wires an output to any input named "<prefix><component>_<port>".
//
// For example, Prefixed("habitat_") wires component gas's environmental_gas
// output to an input named "habitat_gas_environmental_gas".
func Prefixed(prefix string) *Plugin {
	return &Plugin{
		PluginName: "autowire:prefixed:" + prefix,
		InputNameFor: func(source *component.Component, output *port.Port) string {
			return fmt.Sprintf("%s%s_%s", prefix, source.Name(), output.Name())
		},
	}
}

// Broadcast wires every output named portName to every input of the
// same name.
func Broadcast(portName string) *Plugin {
	return BroadcastAs(portName, portName)
}

// BroadcastAs wires every output named outputName to every input named
// inputName.
//
// This is the clock case, where the two differ: a component emitting "tick"
// feeds everything that declared an input called "time", including the
// components added after it.
func BroadcastAs(outputName, inputName string) *Plugin {
	return &Plugin{
		PluginName: "autowire:broadcast:" + outputName + "->" + inputName,
		InputNameFor: func(_ *component.Component, output *port.Port) string {
			if output.Name() != outputName {
				return ""
			}
			return inputName
		},
	}
}

// Name implements fmesh.Plugin.
func (a *Plugin) Name() string {
	if a.PluginName == "" {
		return "autowire"
	}
	return a.PluginName
}

// Init implements fmesh.Plugin.
func (a *Plugin) Init(fm *fmesh.FMesh) error {
	if a.InputNameFor == nil {
		return errors.New("autowire: InputNameFor must be set")
	}

	fm.SetupHooks(func(hooks *fmesh.Hooks) {
		hooks.OnComponentAdded(func(_ context.Context, added *fmesh.ComponentAddedContext) error {
			arrived := added.Component

			for existing := range added.FMesh.Components().All() {
				if existing == arrived {
					// Wiring a component to itself would be a loopback, which is
					// never what a convention meant to express.
					continue
				}
				if err := a.connect(existing, arrived); err != nil {
					return err
				}
				if err := a.connect(arrived, existing); err != nil {
					return err
				}
			}
			return nil
		})
	})
	return nil
}

// connect pipes every output of source to the matching input of destination.
func (a *Plugin) connect(source, destination *component.Component) error {
	for out := range source.Outputs().All() {
		name := a.InputNameFor(source, out)
		if name == "" {
			continue
		}
		in := destination.InputByName(name)
		if in == nil {
			continue
		}
		if isPipedTo(out, in) {
			// Pipes are not deduplicated, and a port flushes once per pipe, so a
			// second identical pipe would deliver every signal twice. Two
			// conventions on the same mesh can easily agree on one pair.
			continue
		}
		if err := out.PipeTo(in); err != nil {
			return err
		}
	}
	return nil
}

// isPipedTo reports whether out already pipes to in.
func isPipedTo(out, in *port.Port) bool {
	return out.Pipes().Find(func(p *port.Port) bool { return p == in }) != nil
}
