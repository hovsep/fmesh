package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/port"
)

type ActivationFunc func(inputs *port.Collection, outputs *port.Collection) error

// Component defines a main building block of FMesh
type Component struct {
	common.NamedEntity
	common.DescribedEntity
	common.LabeledEntity
	*common.Chainable
	inputs  *port.Collection
	outputs *port.Collection
	f       ActivationFunc
}

// New creates initialized component
func New(name string) *Component {
	return &Component{
		NamedEntity:     common.NewNamedEntity(name),
		DescribedEntity: common.NewDescribedEntity(""),
		LabeledEntity:   common.NewLabeledEntity(nil),
		Chainable:       common.NewChainable(),
		inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
			port.DirectionLabel: port.DirectionIn,
		}),
		outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
			port.DirectionLabel: port.DirectionOut,
		}),
	}
}

// WithDescription sets a description
func (c *Component) WithDescription(description string) *Component {
	if c.HasChainError() {
		return c
	}

	c.DescribedEntity = common.NewDescribedEntity(description)
	return c
}

// withInputPorts sets input ports collection
func (c *Component) withInputPorts(collection *port.Collection) *Component {
	if c.HasChainError() {
		return c
	}
	if collection.HasChainError() {
		return c.WithChainError(collection.ChainError())
	}
	c.inputs = collection
	return c
}

// withOutputPorts sets input ports collection
func (c *Component) withOutputPorts(collection *port.Collection) *Component {
	if c.HasChainError() {
		return c
	}
	if collection.HasChainError() {
		return c.WithChainError(collection.ChainError())
	}

	c.outputs = collection
	return c
}

// WithInputs ads input ports
func (c *Component) WithInputs(portNames ...string) *Component {
	if c.HasChainError() {
		return c
	}

	ports, err := port.NewGroup(portNames...).Ports()
	if err != nil {
		c.SetChainError(err)
		return New("").WithChainError(c.ChainError())
	}

	return c.withInputPorts(c.Inputs().With(ports...))
}

// WithOutputs adds output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	if c.HasChainError() {
		return c
	}

	ports, err := port.NewGroup(portNames...).Ports()
	if err != nil {
		c.SetChainError(err)
		return New("").WithChainError(c.ChainError())
	}
	return c.withOutputPorts(c.Outputs().With(ports...))
}

// WithInputsIndexed creates multiple prefixed ports
func (c *Component) WithInputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	if c.HasChainError() {
		return c
	}

	return c.withInputPorts(c.Inputs().WithIndexed(prefix, startIndex, endIndex))
}

// WithOutputsIndexed creates multiple prefixed ports
func (c *Component) WithOutputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	if c.HasChainError() {
		return c
	}

	return c.withOutputPorts(c.Outputs().WithIndexed(prefix, startIndex, endIndex))
}

// WithActivationFunc sets activation function
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	if c.HasChainError() {
		return c
	}

	c.f = f
	return c
}

// WithLabels sets labels and returns the component
func (c *Component) WithLabels(labels common.LabelsCollection) *Component {
	if c.HasChainError() {
		return c
	}
	c.LabeledEntity.SetLabels(labels)
	return c
}

// Inputs getter
func (c *Component) Inputs() *port.Collection {
	if c.HasChainError() {
		return port.NewCollection().WithChainError(c.ChainError())
	}

	return c.inputs
}

// Outputs getter
func (c *Component) Outputs() *port.Collection {
	if c.HasChainError() {
		return port.NewCollection().WithChainError(c.ChainError())
	}

	return c.outputs
}

// OutputByName is shortcut method
func (c *Component) OutputByName(name string) *port.Port {
	if c.HasChainError() {
		return port.New("").WithChainError(c.ChainError())
	}
	outputPort := c.Outputs().ByName(name)
	if outputPort.HasChainError() {
		c.SetChainError(outputPort.ChainError())
		return port.New("").WithChainError(c.ChainError())
	}
	return outputPort
}

// InputByName is shortcut method
func (c *Component) InputByName(name string) *port.Port {
	if c.HasChainError() {
		return port.New("").WithChainError(c.ChainError())
	}
	inputPort := c.Inputs().ByName(name)
	if inputPort.HasChainError() {
		c.SetChainError(inputPort.ChainError())
		return port.New("").WithChainError(c.ChainError())
	}
	return inputPort
}

// hasActivationFunction checks when activation function is set
func (c *Component) hasActivationFunction() bool {
	if c.HasChainError() {
		return false
	}

	return c.f != nil
}

// propagateChainErrors propagates up all chain errors that might have not been propagated yet
func (c *Component) propagateChainErrors() {
	if c.Inputs().HasChainError() {
		c.SetChainError(c.Inputs().ChainError())
		return
	}

	if c.Outputs().HasChainError() {
		c.SetChainError(c.Outputs().ChainError())
		return
	}

	for _, p := range c.Inputs().PortsOrNil() {
		if p.HasChainError() {
			c.SetChainError(p.ChainError())
			return
		}
	}

	for _, p := range c.Outputs().PortsOrNil() {
		if p.HasChainError() {
			c.SetChainError(p.ChainError())
			return
		}
	}
}

// MaybeActivate tries to run the activation function if all required conditions are met
// @TODO: hide this method from user
// @TODO: can we remove named return ?
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	c.propagateChainErrors()

	if c.HasChainError() {
		activationResult = NewActivationResult(c.Name()).WithChainError(c.ChainError())
		return
	}

	defer func() {
		if r := recover(); r != nil {
			activationResult = c.newActivationResultPanicked(fmt.Errorf("panicked with: %v", r))
		}
	}()

	if !c.hasActivationFunction() {
		//Activation function is not set (maybe useful while the mesh is under development)
		activationResult = c.newActivationResultNoFunction()
		return
	}

	if !c.Inputs().AnyHasSignals() {
		//No inputs set, stop here
		activationResult = c.newActivationResultNoInput()
		return
	}

	//Invoke the activation func
	err := c.f(c.Inputs(), c.Outputs())

	if errors.Is(err, errWaitingForInputs) {
		activationResult = c.newActivationResultWaitingForInputs(err)
		return
	}

	if err != nil {
		activationResult = c.newActivationResultReturnedError(err)
		return
	}

	activationResult = c.newActivationResultOK()
	return
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs
func (c *Component) FlushOutputs() *Component {
	if c.HasChainError() {
		return c
	}

	ports, err := c.Outputs().Ports()
	if err != nil {
		c.SetChainError(err)
		return New("").WithChainError(c.ChainError())
	}
	for _, out := range ports {
		out = out.Flush()
		if out.HasChainError() {
			return c.WithChainError(out.ChainError())
		}
	}
	return c
}

// ClearInputs clears all input ports
func (c *Component) ClearInputs() *Component {
	if c.HasChainError() {
		return c
	}
	c.Inputs().Clear()
	return c
}

// WithChainError returns component with error
func (c *Component) WithChainError(err error) *Component {
	c.SetChainError(err)
	return c
}
