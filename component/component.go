package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/hop"
	"github.com/hovsep/fmesh/port"
	"runtime/debug"
)

type ActivationFunc func(inputs port.Ports, outputs port.Ports) error

// Component defines a main building block of FMesh
type Component struct {
	name        string
	description string
	inputs      port.Ports
	outputs     port.Ports
	f           ActivationFunc
}

// Components is a useful collection type
type Components map[string]*Component

// NewComponent creates a new empty component
func NewComponent(name string) *Component {
	return &Component{name: name}
}

func NewComponents(names ...string) Components {
	components := make(Components, len(names))
	for _, name := range names {
		components[name] = NewComponent(name)
	}
	return components
}

// WithDescription sets description
func (c *Component) WithDescription(description string) *Component {
	c.description = description
	return c
}

// WithInputs creates and sets input ports
func (c *Component) WithInputs(portNames ...string) *Component {
	c.inputs = port.NewPorts(portNames...)
	return c
}

// WithOutputs creates and sets output ports
func (c *Component) WithOutputs(portNames ...string) *Component {
	c.outputs = port.NewPorts(portNames...)
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
func (c *Component) Inputs() port.Ports {
	return c.inputs
}

// Outputs getter
func (c *Component) Outputs() port.Ports {
	return c.outputs
}

// Activate runs the activation function
func (c *Component) Activate() (aRes hop.ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			errorFormat := "panicked with: %v, stacktrace: %s"
			if _, ok := r.(error); ok {
				errorFormat = "panicked with: %w, stacktrace: %s"
			}
			aRes = hop.ActivationResult{
				Activated:     true,
				ComponentName: c.name,
				Err:           fmt.Errorf(errorFormat, r, debug.Stack()),
			}
		}
	}()

	//@TODO:: https://github.com/hovsep/fmesh/issues/15
	if !c.inputs.AnyHasSignal() {
		//No inputs set, stop here

		aRes = hop.ActivationResult{
			Activated:     false,
			ComponentName: c.name,
			Err:           nil,
		}

		return
	}

	if c.f == nil {
		//Activation function is not set

		aRes = hop.ActivationResult{
			Activated:     false,
			ComponentName: c.name,
			Err:           nil,
		}

		return
	}

	//Run the computation
	err := c.f(c.inputs, c.outputs)

	if IsWaitingForInputError(err) {
		aRes = hop.ActivationResult{
			Activated:     false,
			ComponentName: c.name,
			Err:           nil,
		}

		if !errors.Is(err, ErrWaitingForInputKeepInputs) {
			c.inputs.ClearSignal()
		}

		return
	}

	//Clear inputs
	c.inputs.ClearSignal()

	if err != nil {
		aRes = hop.ActivationResult{
			Activated:     true,
			ComponentName: c.name,
			Err:           fmt.Errorf("failed to activate component: %w", err),
		}

		return
	}

	aRes = hop.ActivationResult{
		Activated:     true,
		ComponentName: c.name,
		Err:           nil,
	}

	return
}

// FlushOutputs pushed signals out of the component outputs to pipes and clears outputs
func (c *Component) FlushOutputs() {
	for _, out := range c.outputs {
		if !out.HasSignal() || len(out.Pipes()) == 0 {
			continue
		}

		for _, pipe := range out.Pipes() {
			//Multiplexing
			pipe.Flush()
		}
		out.ClearSignal()
	}
}

// ByName returns a component by its name
func (components Components) ByName(name string) *Component {
	return components[name]
}
