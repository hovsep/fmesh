package component

import (
	"fmt"

	"github.com/hovsep/fmesh/port"
)

// withInputPorts sets input ports collection.
// For framework-created ports (via NewGroup), direction is set here.
// For user-created ports (via AttachInputPorts), direction is validated separately.
func (c *Component) withInputPorts(collection *port.Collection) *Component {
	if c.HasChainableErr() {
		return c
	}
	if collection.HasChainableErr() {
		return c.WithChainableErr(collection.ChainableErr())
	}
	c.inputPorts = collection.WithParentComponent(c)
	return c
}

// withOutputPorts sets output ports collection.
// For framework-created ports (via NewGroup), direction is set here.
// For user-created ports (via AttachOutputPorts), direction is validated separately.
func (c *Component) withOutputPorts(collection *port.Collection) *Component {
	if c.HasChainableErr() {
		return c
	}
	if collection.HasChainableErr() {
		return c.WithChainableErr(collection.ChainableErr())
	}
	c.outputPorts = collection.WithParentComponent(c)
	return c
}

// AddInputs creates input ports by name.
func (c *Component) AddInputs(portNames ...string) *Component {
	if c.HasChainableErr() {
		return c
	}

	ports := make([]*port.Port, len(portNames))
	for i, name := range portNames {
		ports[i] = port.NewInput(name)
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// AttachInputPorts attaches pre-configured input ports (must be created with port.NewInput).
func (c *Component) AttachInputPorts(ports ...*port.Port) *Component {
	if c.HasChainableErr() {
		return c
	}

	// Validate that all ports are actually input ports
	for _, p := range ports {
		if !p.IsInput() {
			c.WithChainableErr(fmt.Errorf("AttachInputPorts: port '%s' is not an input port (use port.NewInput): %w", p.Name(), port.ErrWrongPortDirection))
			return New("").WithChainableErr(c.ChainableErr())
		}
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// AddOutputs creates output ports by name.
func (c *Component) AddOutputs(portNames ...string) *Component {
	if c.HasChainableErr() {
		return c
	}

	ports := make([]*port.Port, len(portNames))
	for i, name := range portNames {
		ports[i] = port.NewOutput(name)
	}

	return c.withOutputPorts(c.Outputs().With(ports...))
}

// AttachOutputPorts attaches pre-configured output ports (must be created with port.NewOutput).
func (c *Component) AttachOutputPorts(ports ...*port.Port) *Component {
	if c.HasChainableErr() {
		return c
	}

	// Validate that all ports are actually output ports
	for _, p := range ports {
		if !p.IsOutput() {
			c.WithChainableErr(fmt.Errorf("AttachOutputPorts: port '%s' is not an output port (use port.NewOutput): %w", p.Name(), port.ErrWrongPortDirection))
			return New("").WithChainableErr(c.ChainableErr())
		}
	}

	return c.withOutputPorts(c.Outputs().With(ports...))
}

// WithInputsIndexed creates multiple prefixed input ports.
func (c *Component) WithInputsIndexed(prefix string, startIndex, endIndex int) *Component {
	if c.HasChainableErr() {
		return c
	}

	if startIndex > endIndex {
		c.WithChainableErr(port.ErrInvalidRangeForIndexedGroup)
		return New("").WithChainableErr(c.ChainableErr())
	}

	ports := make([]*port.Port, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		ports = append(ports, port.NewInput(fmt.Sprintf("%s%d", prefix, i)))
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// WithOutputsIndexed creates multiple prefixed output ports.
func (c *Component) WithOutputsIndexed(prefix string, startIndex, endIndex int) *Component {
	if c.HasChainableErr() {
		return c
	}

	if startIndex > endIndex {
		c.WithChainableErr(port.ErrInvalidRangeForIndexedGroup)
		return New("").WithChainableErr(c.ChainableErr())
	}

	ports := make([]*port.Port, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		ports = append(ports, port.NewOutput(fmt.Sprintf("%s%d", prefix, i)))
	}

	return c.withOutputPorts(c.Outputs().With(ports...))
}

// Inputs returns the component's input ports.
func (c *Component) Inputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.inputPorts
}

// Outputs returns the component's output ports.
func (c *Component) Outputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.outputPorts
}

// OutputByName returns an output port by name.
func (c *Component) OutputByName(name string) *port.Port {
	if c.HasChainableErr() {
		return port.NewOutput("").WithChainableErr(c.ChainableErr())
	}
	outputPort := c.Outputs().ByName(name)
	if outputPort.HasChainableErr() {
		c.WithChainableErr(outputPort.ChainableErr())
		return port.NewOutput("").WithChainableErr(c.ChainableErr())
	}
	return outputPort
}

// InputByName returns an input port by name.
func (c *Component) InputByName(name string) *port.Port {
	if c.HasChainableErr() {
		return port.NewInput("").WithChainableErr(c.ChainableErr())
	}
	inputPort := c.Inputs().ByName(name)
	if inputPort.HasChainableErr() {
		c.WithChainableErr(inputPort.ChainableErr())
		return port.NewInput("").WithChainableErr(c.ChainableErr())
	}
	return inputPort
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs.
func (c *Component) FlushOutputs() *Component {
	if c.HasChainableErr() {
		return c
	}

	ports, err := c.Outputs().All()
	if err != nil {
		c.WithChainableErr(err)
		return New("").WithChainableErr(c.ChainableErr())
	}
	for _, out := range ports {
		out = out.Flush()
		if out.HasChainableErr() {
			return c.WithChainableErr(out.ChainableErr())
		}
	}
	return c
}

// ClearInputs clears all input ports.
func (c *Component) ClearInputs() *Component {
	if c.HasChainableErr() {
		return c
	}
	c.Inputs().ForEach(func(p *port.Port) {
		p.Clear()
	})
	return c
}

// LoopbackPipe creates a pipe between ports of the component.
func (c *Component) LoopbackPipe(out, in string) {
	c.OutputByName(out).PipeTo(c.InputByName(in))
}
