package component

import (
	"context"
	"fmt"

	"github.com/hovsep/fmesh/port"
)

// WithInputs is a component option that creates named input ports.
func WithInputs(portNames ...string) Option {
	return func(c *Component) error {
		return c.addPorts(port.DirectionIn, portNames...)
	}
}

// WithOutputs is a component option that creates named output ports.
func WithOutputs(portNames ...string) Option {
	return func(c *Component) error {
		return c.addPorts(port.DirectionOut, portNames...)
	}
}

// WithIndexedInputs is a component option that creates prefixed indexed input ports.
func WithIndexedInputs(prefix string, startIndex, endIndex int) Option {
	return func(c *Component) error {
		return c.addIndexedPorts(port.DirectionIn, prefix, startIndex, endIndex)
	}
}

// WithIndexedOutputs is a component option that creates prefixed indexed output ports.
func WithIndexedOutputs(prefix string, startIndex, endIndex int) Option {
	return func(c *Component) error {
		return c.addIndexedPorts(port.DirectionOut, prefix, startIndex, endIndex)
	}
}

// portSide is everything about a component's I/O that depends on which side of it
// you are on, so the input and output paths can be one implementation instead of
// two near-identical ones.
type portSide struct {
	ports    *port.Collection
	create   func(string, ...port.Option) (*port.Port, error)
	name     string // "input" or "output", for error messages
	ctorName string // the constructor to point a caller at, for error messages
}

func (c *Component) sideFor(direction port.Direction) portSide {
	if direction == port.DirectionIn {
		return portSide{c.inputPorts, port.NewInput, "input", "port.NewInput"}
	}
	return portSide{c.outputPorts, port.NewOutput, "output", "port.NewOutput"}
}

// addPorts creates ports by name on one side of the component.
func (c *Component) addPorts(direction port.Direction, portNames ...string) error {
	side := c.sideFor(direction)
	ports := make([]*port.Port, 0, len(portNames))
	for _, name := range portNames {
		p, err := side.create(name)
		if err != nil {
			return fmt.Errorf("failed to create %s port %q: %w", side.name, name, err)
		}
		ports = append(ports, p)
	}
	return c.attach(side, ports, "add")
}

// addIndexedPorts creates prefixed ports numbered startIndex..endIndex inclusive.
func (c *Component) addIndexedPorts(direction port.Direction, prefix string, startIndex, endIndex int) error {
	if startIndex > endIndex {
		return port.ErrInvalidRangeForIndexedGroup
	}

	side := c.sideFor(direction)
	ports := make([]*port.Port, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		p, err := side.create(fmt.Sprintf("%s%d", prefix, i))
		if err != nil {
			return fmt.Errorf("failed to create indexed %s port: %w", side.name, err)
		}
		ports = append(ports, p)
	}
	return c.attach(side, ports, "add indexed")
}

// attach adds already-built ports to one side and adopts them. verb names what the
// caller was doing, and is a constant so the message costs nothing unless Add
// fails — building it eagerly cost an allocation on every port added.
func (c *Component) attach(side portSide, ports []*port.Port, verb string) error {
	if err := side.ports.Add(ports...); err != nil {
		return fmt.Errorf("failed to %s %s ports: %w", verb, side.name, err)
	}
	side.ports.SetParentComponent(c)
	return nil
}

// attachPorts adds pre-built ports, rejecting any facing the wrong way.
func (c *Component) attachPorts(direction port.Direction, ports ...*port.Port) error {
	side := c.sideFor(direction)
	for _, p := range ports {
		if p.Direction() != direction {
			return fmt.Errorf("cannot attach %s port %q: it is not an %s port (create it with %s): %w",
				side.name, p.Name(), side.name, side.ctorName, port.ErrWrongPortDirection)
		}
	}
	return c.attach(side, ports, "attach")
}

// AddInputs creates input ports by name and attaches them to the component.
// Returns an error if any port cannot be created.
func (c *Component) AddInputs(portNames ...string) error {
	return c.addPorts(port.DirectionIn, portNames...)
}

// AddOutputs creates output ports by name and attaches them to the component.
// Returns an error if any port cannot be created.
func (c *Component) AddOutputs(portNames ...string) error {
	return c.addPorts(port.DirectionOut, portNames...)
}

// AddIndexedInputs creates multiple prefixed input ports and attaches them to the component.
func (c *Component) AddIndexedInputs(prefix string, startIndex, endIndex int) error {
	return c.addIndexedPorts(port.DirectionIn, prefix, startIndex, endIndex)
}

// AddIndexedOutputs creates multiple prefixed output ports and attaches them to the component.
func (c *Component) AddIndexedOutputs(prefix string, startIndex, endIndex int) error {
	return c.addIndexedPorts(port.DirectionOut, prefix, startIndex, endIndex)
}

// AttachInputPorts attaches pre-configured input ports (must be created with port.NewInput).
func (c *Component) AttachInputPorts(ports ...*port.Port) error {
	return c.attachPorts(port.DirectionIn, ports...)
}

// AttachOutputPorts attaches pre-configured output ports (must be created with port.NewOutput).
func (c *Component) AttachOutputPorts(ports ...*port.Port) error {
	return c.attachPorts(port.DirectionOut, ports...)
}

// Inputs returns the component's input ports.
func (c *Component) Inputs() *port.Collection {
	return c.inputPorts
}

// Outputs returns the component's output ports.
func (c *Component) Outputs() *port.Collection {
	return c.outputPorts
}

// OutputByName returns an output port by name.
func (c *Component) OutputByName(name string) *port.Port {
	return c.Outputs().ByName(name)
}

// InputByName returns an input port by name.
func (c *Component) InputByName(name string) *port.Port {
	return c.Inputs().ByName(name)
}

// FlushOutputs pushes signals out of the component outputs to pipes and clears outputs.
func (c *Component) FlushOutputs(ctx context.Context) error {
	return c.Outputs().ForEach(func(out *port.Port) error {
		if err := out.Flush(ctx); err != nil {
			return fmt.Errorf("failed to flush output port %q: %w", out.Name(), err)
		}
		return nil
	})
}

// ClearInputs clears all input ports.
func (c *Component) ClearInputs(ctx context.Context) error {
	return c.Inputs().ForEach(func(p *port.Port) error {
		return p.Clear(ctx)
	})
}

// ClearOutputs clears all output ports.
func (c *Component) ClearOutputs(ctx context.Context) error {
	return c.Outputs().ForEach(func(p *port.Port) error {
		return p.Clear(ctx)
	})
}

// LoopbackPipe creates a pipe between output and input ports of the component.
func (c *Component) LoopbackPipe(out, in string) error {
	outPort := c.OutputByName(out)
	if outPort == nil {
		return fmt.Errorf("LoopbackPipe: output port %q not found", out)
	}
	inPort := c.InputByName(in)
	if inPort == nil {
		return fmt.Errorf("LoopbackPipe: input port %q not found", in)
	}
	return outPort.PipeTo(inPort)
}
