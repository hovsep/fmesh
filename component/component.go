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
		inputs:          port.NewCollection(),
		outputs:         port.NewCollection(),
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

// WithInputs ads input ports
func (c *Component) WithInputs(portNames ...string) *Component {
	if c.HasChainError() {
		return c
	}

	ports, err := port.NewGroup(portNames...).Ports()
	if err != nil {
		return c.WithChainError(err)
	}
	c.inputs = c.Inputs().With(ports...)
	return c
}

// WithOutputs adds output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	if c.HasChainError() {
		return c
	}
	ports, err := port.NewGroup(portNames...).Ports()
	if err != nil {
		return c.WithChainError(err)
	}
	c.outputs = c.Outputs().With(ports...)
	return c
}

// WithInputsIndexed creates multiple prefixed ports
func (c *Component) WithInputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	if c.HasChainError() {
		return c
	}

	c.inputs = c.Inputs().WithIndexed(prefix, startIndex, endIndex)
	return c
}

// WithOutputsIndexed creates multiple prefixed ports
func (c *Component) WithOutputsIndexed(prefix string, startIndex int, endIndex int) *Component {
	if c.HasChainError() {
		return c
	}

	c.outputs = c.Outputs().WithIndexed(prefix, startIndex, endIndex)
	return c
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
	outputPort := c.Outputs().ByName(name)
	if outputPort.HasChainError() {
		c.SetChainError(outputPort.ChainError())
		return nil
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

// MaybeActivate tries to run the activation function if all required conditions are met
// @TODO: hide this method from user
// @TODO: can we remove named return ?
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	//Bubble up chain errors from ports
	for _, p := range c.Inputs().PortsOrNil() {
		if p.HasChainError() {
			c.Inputs().SetChainError(p.ChainError())
			c.SetChainError(c.Inputs().ChainError())
			break
		}
	}
	for _, p := range c.Outputs().PortsOrNil() {
		if p.HasChainError() {
			c.Outputs().SetChainError(p.ChainError())
			c.SetChainError(c.Outputs().ChainError())
			break
		}
	}

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

	if !c.inputs.AnyHasSignals() {
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

	ports, err := c.outputs.Ports()
	if err != nil {
		return c.WithChainError(err)
	}
	for _, out := range ports {
		out.Flush()
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
