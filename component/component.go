package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/port"
	"log"
)

type ActivationFunc func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error

// Component defines a main building block of FMesh
type Component struct {
	common.NamedEntity
	common.DescribedEntity
	common.LabeledEntity
	*common.Chainable
	inputs  *port.Collection
	outputs *port.Collection
	f       ActivationFunc
	logger  *log.Logger
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
	if c.HasErr() {
		return c
	}

	c.DescribedEntity = common.NewDescribedEntity(description)
	return c
}

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

// WithActivationFunc sets activation function
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	if c.HasErr() {
		return c
	}

	c.f = f
	return c
}

// WithLabels sets labels and returns the component
func (c *Component) WithLabels(labels common.LabelsCollection) *Component {
	if c.HasErr() {
		return c
	}
	c.LabeledEntity.SetLabels(labels)
	return c
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

// hasActivationFunction checks when activation function is set
func (c *Component) hasActivationFunction() bool {
	if c.HasErr() {
		return false
	}

	return c.f != nil
}

// propagateChainErrors propagates up all chain errors that might have not been propagated yet
func (c *Component) propagateChainErrors() {
	if c.Inputs().HasErr() {
		c.SetErr(c.Inputs().Err())
		return
	}

	if c.Outputs().HasErr() {
		c.SetErr(c.Outputs().Err())
		return
	}

	for _, p := range c.Inputs().PortsOrNil() {
		if p.HasErr() {
			c.SetErr(p.Err())
			return
		}
	}

	for _, p := range c.Outputs().PortsOrNil() {
		if p.HasErr() {
			c.SetErr(p.Err())
			return
		}
	}
}

// MaybeActivate tries to run the activation function if all required conditions are met
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	c.propagateChainErrors()

	if c.HasErr() {
		activationResult = NewActivationResult(c.Name()).WithErr(c.Err())
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
	err := c.f(c.Inputs(), c.Outputs(), c.Logger())

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

// WithErr returns component with error
func (c *Component) WithErr(err error) *Component {
	c.SetErr(err)
	return c
}

// WithPrefixedLogger creates a new logger prefixed with component name
func (c *Component) WithPrefixedLogger(logger *log.Logger) *Component {
	if c.HasErr() {
		return c
	}

	if logger == nil {
		return c
	}

	prefix := fmt.Sprintf("%s : ", c.Name())
	c.logger = log.New(logger.Writer(), prefix, logger.Flags())
	return c
}

func (c *Component) Logger() *log.Logger {
	return c.logger
}
