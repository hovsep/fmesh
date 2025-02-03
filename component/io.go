package component

import "github.com/hovsep/fmesh/port"

// withInputPorts sets input ports collection
func (c *Component) withInputPorts(collection *port.Collection) *Component {
	if c.HasErr() {
		return c
	}
	if collection.HasErr() {
		return c.WithErr(collection.Err())
	}
	c.inputs = collection
	return c
}

// withOutputPorts sets input ports collection
func (c *Component) withOutputPorts(collection *port.Collection) *Component {
	if c.HasErr() {
		return c
	}
	if collection.HasErr() {
		return c.WithErr(collection.Err())
	}

	c.outputs = collection
	return c
}

// WithInputs ads input ports
func (c *Component) WithInputs(portNames ...string) *Component {
	if c.HasErr() {
		return c
	}

	ports, err := port.NewGroup(portNames...).Ports()
	if err != nil {
		c.SetErr(err)
		return New("").WithErr(c.Err())
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// WithOutputs adds output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	if c.HasErr() {
		return c
	}

	ports, err := port.NewGroup(portNames...).Ports()
	if err != nil {
		c.SetErr(err)
		return New("").WithErr(c.Err())
	}
	return c.withOutputPorts(c.Outputs().With(ports...))
}

// WithInputsIndexed creates multiple prefixed ports
func (c *Component) WithInputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	if c.HasErr() {
		return c
	}

	return c.withInputPorts(c.Inputs().WithIndexed(prefix, startIndex, endIndex))
}

// WithOutputsIndexed creates multiple prefixed ports
func (c *Component) WithOutputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	if c.HasErr() {
		return c
	}

	return c.withOutputPorts(c.Outputs().WithIndexed(prefix, startIndex, endIndex))
}

// Inputs getter
func (c *Component) Inputs() *port.Collection {
	if c.HasErr() {
		return port.NewCollection().WithErr(c.Err())
	}

	return c.inputs
}

// Outputs getter
func (c *Component) Outputs() *port.Collection {
	if c.HasErr() {
		return port.NewCollection().WithErr(c.Err())
	}

	return c.outputs
}

// OutputByName is shortcut method
func (c *Component) OutputByName(name string) *port.Port {
	if c.HasErr() {
		return port.New("").WithErr(c.Err())
	}
	outputPort := c.Outputs().ByName(name)
	if outputPort.HasErr() {
		c.SetErr(outputPort.Err())
		return port.New("").WithErr(c.Err())
	}
	return outputPort
}

// InputByName is shortcut method
func (c *Component) InputByName(name string) *port.Port {
	if c.HasErr() {
		return port.New("").WithErr(c.Err())
	}
	inputPort := c.Inputs().ByName(name)
	if inputPort.HasErr() {
		c.SetErr(inputPort.Err())
		return port.New("").WithErr(c.Err())
	}
	return inputPort
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs
func (c *Component) FlushOutputs() *Component {
	if c.HasErr() {
		return c
	}

	ports, err := c.Outputs().Ports()
	if err != nil {
		c.SetErr(err)
		return New("").WithErr(c.Err())
	}
	for _, out := range ports {
		out = out.Flush()
		if out.HasErr() {
			return c.WithErr(out.Err())
		}
	}
	return c
}

// ClearInputs clears all input ports
func (c *Component) ClearInputs() *Component {
	if c.HasErr() {
		return c
	}
	c.Inputs().Clear()
	return c
}
