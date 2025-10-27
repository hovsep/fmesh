package component

import "github.com/hovsep/fmesh/port"

// withInputPorts sets input ports collection.
func (c *Component) withInputPorts(collection *port.Collection) *Component {
	if c.HasChainableErr() {
		return c
	}
	if collection.HasChainableErr() {
		return c.WithChainableErr(collection.ChainableErr())
	}
	c.inputs = collection.WithParentComponent(c)
	return c
}

// withOutputPorts sets input ports collection.
func (c *Component) withOutputPorts(collection *port.Collection) *Component {
	if c.HasChainableErr() {
		return c
	}
	if collection.HasChainableErr() {
		return c.WithChainableErr(collection.ChainableErr())
	}

	c.outputs = collection.WithParentComponent(c)
	return c
}

// WithInputs ads input ports.
func (c *Component) WithInputs(portNames ...string) *Component {
	if c.HasChainableErr() {
		return c
	}

	ports, err := port.NewGroup(portNames...).AllAsSlice()
	if err != nil {
		c.WithChainableErr(err)
		return New("").WithChainableErr(c.ChainableErr())
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// WithOutputs adds output ports.
func (c *Component) WithOutputs(portNames ...string) *Component {
	if c.HasChainableErr() {
		return c
	}

	ports, err := port.NewGroup(portNames...).AllAsSlice()
	if err != nil {
		c.WithChainableErr(err)
		return New("").WithChainableErr(c.ChainableErr())
	}
	return c.withOutputPorts(c.Outputs().With(ports...))
}

// WithInputsIndexed creates multiple prefixed ports.
func (c *Component) WithInputsIndexed(prefix string, startIndex, endIndex int) *Component {
	if c.HasChainableErr() {
		return c
	}

	return c.withInputPorts(c.Inputs().WithIndexed(prefix, startIndex, endIndex))
}

// WithOutputsIndexed creates multiple prefixed ports.
func (c *Component) WithOutputsIndexed(prefix string, startIndex, endIndex int) *Component {
	if c.HasChainableErr() {
		return c
	}

	return c.withOutputPorts(c.Outputs().WithIndexed(prefix, startIndex, endIndex))
}

// Inputs getter.
func (c *Component) Inputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.inputs
}

// Outputs getter.
func (c *Component) Outputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.outputs
}

// OutputByName is shortcut method.
func (c *Component) OutputByName(name string) *port.Port {
	if c.HasChainableErr() {
		return port.New("").WithChainableErr(c.ChainableErr())
	}
	outputPort := c.Outputs().ByName(name)
	if outputPort.HasChainableErr() {
		c.WithChainableErr(outputPort.ChainableErr())
		return port.New("").WithChainableErr(c.ChainableErr())
	}
	return outputPort
}

// InputByName is shortcut method.
func (c *Component) InputByName(name string) *port.Port {
	if c.HasChainableErr() {
		return port.New("").WithChainableErr(c.ChainableErr())
	}
	inputPort := c.Inputs().ByName(name)
	if inputPort.HasChainableErr() {
		c.WithChainableErr(inputPort.ChainableErr())
		return port.New("").WithChainableErr(c.ChainableErr())
	}
	return inputPort
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs.
func (c *Component) FlushOutputs() *Component {
	if c.HasChainableErr() {
		return c
	}

	ports, err := c.Outputs().AllAsSlice()
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
	c.Inputs().Clear()
	return c
}

// LoopbackPipe creates a pipe between ports of the component.
func (c *Component) LoopbackPipe(out, in string) {
	c.OutputByName(out).PipeTo(c.InputByName(in))
}
