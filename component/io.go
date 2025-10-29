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
	c.inputPorts = collection.WithParentComponent(c)
	return c
}

// withOutputPorts sets output ports collection.
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

// AddInputs creates and adds input ports by name (simple API).
// Use this when you only need to specify port names.
// For ports with descriptions, labels, or other configuration, use AttachInputPorts.
//
// Example:
//
//	c := component.New("processor").
//	    AddInputs("data", "config", "metadata")
func (c *Component) AddInputs(portNames ...string) *Component {
	if c.HasChainableErr() {
		return c
	}

	ports, err := port.NewGroup(portNames...).All()
	if err != nil {
		c.WithChainableErr(err)
		return New("").WithChainableErr(c.ChainableErr())
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// AttachInputPorts attaches pre-configured input port instances (advanced API).
// Use this when you need to set descriptions, labels, or other port configuration.
// Can be mixed with AddInputs for flexibility.
//
// Example:
//
//	c := component.New("processor").
//	    AttachInputPorts(
//	        port.New("request").
//	            WithDescription("HTTP request data").
//	            AddLabel("content-type", "json"),
//	    )
func (c *Component) AttachInputPorts(ports ...*port.Port) *Component {
	if c.HasChainableErr() {
		return c
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// AddOutputs creates and adds output ports by name (simple API).
// Use this when you only need to specify port names.
// For ports with descriptions, labels, or other configuration, use AttachOutputPorts.
//
// Example:
//
//	c := component.New("processor").
//	    AddOutputs("result", "error", "logs")
func (c *Component) AddOutputs(portNames ...string) *Component {
	if c.HasChainableErr() {
		return c
	}

	ports, err := port.NewGroup(portNames...).All()
	if err != nil {
		c.WithChainableErr(err)
		return New("").WithChainableErr(c.ChainableErr())
	}
	return c.withOutputPorts(c.Outputs().With(ports...))
}

// AttachOutputPorts attaches pre-configured output port instances (advanced API).
// Use this when you need to set descriptions, labels, or other port configuration.
// Can be mixed with AddOutputs for flexibility.
//
// Example:
//
//	c := component.New("processor").
//	    AttachOutputPorts(
//	        port.New("response").
//	            WithDescription("HTTP response data").
//	            AddLabel("status", "success"),
//	        port.New("error").
//	            WithDescription("Error details").
//	            AddLabel("status", "error"),
//	    )
func (c *Component) AttachOutputPorts(ports ...*port.Port) *Component {
	if c.HasChainableErr() {
		return c
	}

	return c.withOutputPorts(c.Outputs().With(ports...))
}

// WithInputs is deprecated. Use AddInputs instead.
func (c *Component) WithInputs(portNames ...string) *Component {
	return c.AddInputs(portNames...)
}

// WithInputPorts is deprecated. Use AttachInputPorts instead.
func (c *Component) WithInputPorts(ports ...*port.Port) *Component {
	return c.AttachInputPorts(ports...)
}

// WithOutputs is deprecated. Use AddOutputs instead.
func (c *Component) WithOutputs(portNames ...string) *Component {
	return c.AddOutputs(portNames...)
}

// WithOutputPorts is deprecated. Use AttachOutputPorts instead.
func (c *Component) WithOutputPorts(ports ...*port.Port) *Component {
	return c.AttachOutputPorts(ports...)
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

// Inputs returns the component's input ports collection.
func (c *Component) Inputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.inputPorts
}

// Outputs returns the component's output ports collection.
func (c *Component) Outputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.outputPorts
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
