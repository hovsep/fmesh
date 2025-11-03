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

	ports := make([]*port.Port, len(portNames))
	for i, name := range portNames {
		ports[i] = port.NewInput(name)
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// AttachInputPorts attaches pre-configured input port instances (advanced API).
// Use this when you need to set descriptions, labels, or other port configuration.
// Can be mixed with AddInputs for flexibility.
//
// Important: Ports must be created with port.NewInput(). Ports with wrong direction will cause an error.
//
// Example:
//
//	c := component.New("processor").
//	    AttachInputPorts(
//	        port.NewInput("request").
//	            WithDescription("HTTP request data").
//	            AddLabel("content-type", "json"),
//	    )
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

	ports := make([]*port.Port, len(portNames))
	for i, name := range portNames {
		ports[i] = port.NewOutput(name)
	}

	return c.withOutputPorts(c.Outputs().With(ports...))
}

// AttachOutputPorts attaches pre-configured output port instances (advanced API).
// Use this when you need to set descriptions, labels, or other port configuration.
// Can be mixed with AddOutputs for flexibility.
//
// Important: Ports must be created with port.NewOutput(). Ports with wrong direction will cause an error.
//
// Example:
//
//	c := component.New("processor").
//	    AttachOutputPorts(
//	        port.NewOutput("response").
//	            WithDescription("HTTP response data").
//	            AddLabel("status", "success"),
//	        port.NewOutput("error").
//	            WithDescription("Error details").
//	            AddLabel("status", "error"),
//	    )
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

// Inputs returns the component's input ports collection.
// Use this to access multiple ports or perform collection operations like filtering.
//
// Example (in activation function):
//
//	if !this.Inputs().AllHaveSignals() {
//	    return nil // Wait for all inputs
//	}
func (c *Component) Inputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.inputPorts
}

// Outputs returns the component's output ports collection.
// Use this to access multiple output ports or perform collection operations.
//
// Example (in activation function):
//
//	this.Outputs().ForEach(func(p *port.Port) {
//	    p.Clear() // Clear all outputs
//	})
func (c *Component) Outputs() *port.Collection {
	if c.HasChainableErr() {
		return port.NewCollection().WithChainableErr(c.ChainableErr())
	}

	return c.outputPorts
}

// OutputByName retrieves an output port by its name.
// This is the most common way to write data from your component.
//
// Example (in activation function):
//
//	result := processData(input)
//	this.OutputByName("result").PutSignals(signal.New(result))
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

// InputByName retrieves an input port by its name.
// This is the most common way to read data in your component's activation function.
//
// Example (in activation function):
//
//	data := this.InputByName("data").Signals().FirstPayloadOrDefault("").(string)
//	config := this.InputByName("config").Signals().FirstPayloadOrNil()
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
