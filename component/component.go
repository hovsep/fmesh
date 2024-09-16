package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/port"
)

type ActivationFunc func(inputs port.Collection, outputs port.Collection) error

// Component defines a main building block of FMesh
type Component struct {
	name        string
	description string
	inputs      port.Collection
	outputs     port.Collection
	f           ActivationFunc
}

// Components is a useful collection type
type Components map[string]*Component

// NewComponent creates a new empty component
// TODO: rename all constructors to New
func NewComponent(name string) *Component {
	return &Component{name: name}
}

// NewComponents creates a collection of components
// names are optional and can be used to create multiple empty components in one call
// @TODO: rename all such constructors to NewCollection
func NewComponents(names ...string) Components {
	components := make(Components, len(names))
	for _, name := range names {
		components[name] = NewComponent(name)
	}
	return components
}

// WithDescription sets a description
func (c *Component) WithDescription(description string) *Component {
	c.description = description
	return c
}

// WithInputs creates input ports
func (c *Component) WithInputs(portNames ...string) *Component {
	c.inputs = port.NewPortsCollection().Add(port.NewPortGroup(portNames...)...)
	return c
}

// WithOutputs creates output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	c.outputs = port.NewPortsCollection().Add(port.NewPortGroup(portNames...)...)
	return c
}

// WithActivationFunc sets activation function
func (c *Component) WithActivationFunc(f ActivationFunc) *Component {
	c.f = f
	return c
}

// Name getter
func (c *Component) Name() string {
	return c.name
}

// Description getter
func (c *Component) Description() string {
	return c.description
}

// Inputs getter
func (c *Component) Inputs() port.Collection {
	return c.inputs
}

// Outputs getter
func (c *Component) Outputs() port.Collection {
	return c.outputs
}

// hasActivationFunction checks when activation function is set
func (c *Component) hasActivationFunction() bool {
	return c.f != nil
}

// MaybeActivate tries to run the activation function if all required conditions are met
func (c *Component) MaybeActivate() (activationResult *ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			//Clear inputs and exit
			c.inputs.ClearSignal()
			activationResult = c.newActivationCodePanicked(fmt.Errorf("panicked with: %v", r))
		}
	}()

	if !c.hasActivationFunction() {
		//Activation function is not set (maybe useful while the mesh is under development)
		activationResult = c.newActivationCodeNoFunction()

		return
	}

	//@TODO:: https://github.com/hovsep/fmesh/issues/15
	if !c.inputs.AnyHasSignal() {
		//No inputs set, stop here
		activationResult = c.newActivationCodeNoInput()

		return
	}

	//Run the computation
	err := c.f(c.inputs, c.outputs)

	if IsWaitingForInputError(err) {
		activationResult = c.newActivationCodeWaitingForInput()

		if !errors.Is(err, ErrWaitingForInputKeepInputs) {
			c.inputs.ClearSignal()
		}

		return
	}

	//Clear inputs
	c.inputs.ClearSignal()

	if err != nil {
		activationResult = c.newActivationCodeReturnedError(err)

		return
	}

	activationResult = c.newActivationResultOK()

	return
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs
func (c *Component) FlushOutputs() {
	for _, out := range c.outputs {
		out.Flush()
	}
}

// ByName returns a component by its name
func (components Components) ByName(name string) *Component {
	return components[name]
}

// Add adds new components to existing collection
func (components Components) Add(newComponents ...*Component) Components {
	for _, component := range newComponents {
		components[component.Name()] = component
	}
	return components
}
