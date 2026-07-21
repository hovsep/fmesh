package component

import (
	"fmt"

	"github.com/hovsep/fmesh/port"
)

// WithInputs is a component option that creates named input ports.
func WithInputs(portNames ...string) Option {
	return func(c *Component) error {
		return c.addInputs(portNames...)
	}
}

// WithOutputs is a component option that creates named output ports.
func WithOutputs(portNames ...string) Option {
	return func(c *Component) error {
		return c.addOutputs(portNames...)
	}
}

// WithIndexedInputs is a component option that creates prefixed indexed input ports.
func WithIndexedInputs(prefix string, startIndex, endIndex int) Option {
	return func(c *Component) error {
		return c.addIndexedInputs(prefix, startIndex, endIndex)
	}
}

// WithIndexedOutputs is a component option that creates prefixed indexed output ports.
func WithIndexedOutputs(prefix string, startIndex, endIndex int) Option {
	return func(c *Component) error {
		return c.addIndexedOutputs(prefix, startIndex, endIndex)
	}
}

// addInputs creates input ports by name and adds them to the component.
func (c *Component) addInputs(portNames ...string) error {
	ports := make([]*port.Port, 0, len(portNames))
	for _, name := range portNames {
		p, err := port.NewInput(name)
		if err != nil {
			return fmt.Errorf("failed to create input port %q: %w", name, err)
		}
		ports = append(ports, p)
	}
	if err := c.inputPorts.Add(ports...); err != nil {
		return fmt.Errorf("failed to add input ports: %w", err)
	}
	c.inputPorts.WithParentComponent(c)
	return nil
}

// addOutputs creates output ports by name and adds them to the component.
func (c *Component) addOutputs(portNames ...string) error {
	ports := make([]*port.Port, 0, len(portNames))
	for _, name := range portNames {
		p, err := port.NewOutput(name)
		if err != nil {
			return fmt.Errorf("failed to create output port %q: %w", name, err)
		}
		ports = append(ports, p)
	}
	if err := c.outputPorts.Add(ports...); err != nil {
		return fmt.Errorf("failed to add output ports: %w", err)
	}
	c.outputPorts.WithParentComponent(c)
	return nil
}

// addIndexedInputs creates multiple prefixed input ports.
func (c *Component) addIndexedInputs(prefix string, startIndex, endIndex int) error {
	if startIndex > endIndex {
		return port.ErrInvalidRangeForIndexedGroup
	}

	ports := make([]*port.Port, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		p, err := port.NewInput(fmt.Sprintf("%s%d", prefix, i))
		if err != nil {
			return fmt.Errorf("failed to create indexed input port: %w", err)
		}
		ports = append(ports, p)
	}
	if err := c.inputPorts.Add(ports...); err != nil {
		return fmt.Errorf("failed to add indexed input ports: %w", err)
	}
	c.inputPorts.WithParentComponent(c)
	return nil
}

// addIndexedOutputs creates multiple prefixed output ports.
func (c *Component) addIndexedOutputs(prefix string, startIndex, endIndex int) error {
	if startIndex > endIndex {
		return port.ErrInvalidRangeForIndexedGroup
	}

	ports := make([]*port.Port, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		p, err := port.NewOutput(fmt.Sprintf("%s%d", prefix, i))
		if err != nil {
			return fmt.Errorf("failed to create indexed output port: %w", err)
		}
		ports = append(ports, p)
	}
	if err := c.outputPorts.Add(ports...); err != nil {
		return fmt.Errorf("failed to add indexed output ports: %w", err)
	}
	c.outputPorts.WithParentComponent(c)
	return nil
}

// AddInputs creates input ports by name and attaches them to the component.
// Returns an error if any port cannot be created.
func (c *Component) AddInputs(portNames ...string) error {
	return c.addInputs(portNames...)
}

// AddOutputs creates output ports by name and attaches them to the component.
// Returns an error if any port cannot be created.
func (c *Component) AddOutputs(portNames ...string) error {
	return c.addOutputs(portNames...)
}

// AddIndexedInputs creates multiple prefixed input ports and attaches them to the component.
func (c *Component) AddIndexedInputs(prefix string, startIndex, endIndex int) error {
	return c.addIndexedInputs(prefix, startIndex, endIndex)
}

// AddIndexedOutputs creates multiple prefixed output ports and attaches them to the component.
func (c *Component) AddIndexedOutputs(prefix string, startIndex, endIndex int) error {
	return c.addIndexedOutputs(prefix, startIndex, endIndex)
}

// AttachInputPorts attaches pre-configured input ports (must be created with port.NewInput).
func (c *Component) AttachInputPorts(ports ...*port.Port) error {
	for _, p := range ports {
		if !p.IsInput() {
			return fmt.Errorf("AttachInputPorts: port %q is not an input port (use port.NewInput): %w", p.Name(), port.ErrWrongPortDirection)
		}
	}
	if err := c.inputPorts.Add(ports...); err != nil {
		return fmt.Errorf("AttachInputPorts: %w", err)
	}
	c.inputPorts.WithParentComponent(c)
	return nil
}

// AttachOutputPorts attaches pre-configured output ports (must be created with port.NewOutput).
func (c *Component) AttachOutputPorts(ports ...*port.Port) error {
	for _, p := range ports {
		if !p.IsOutput() {
			return fmt.Errorf("AttachOutputPorts: port %q is not an output port (use port.NewOutput): %w", p.Name(), port.ErrWrongPortDirection)
		}
	}
	if err := c.outputPorts.Add(ports...); err != nil {
		return fmt.Errorf("AttachOutputPorts: %w", err)
	}
	c.outputPorts.WithParentComponent(c)
	return nil
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
func (c *Component) FlushOutputs() error {
	return c.Outputs().ForEach(func(out *port.Port) error {
		if err := out.Flush(); err != nil {
			return fmt.Errorf("failed to flush output port %q: %w", out.Name(), err)
		}
		return nil
	})
}

// ClearInputs clears all input ports.
func (c *Component) ClearInputs() error {
	return c.Inputs().ForEach(func(p *port.Port) error {
		return p.Clear()
	})
}

// ClearOutputs clears all output ports.
func (c *Component) ClearOutputs() error {
	return c.Outputs().ForEach(func(p *port.Port) error {
		return p.Clear()
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
