package component

import (
	"errors"
	"fmt"
	"github.com/hovsep/fmesh/hop"
	"github.com/hovsep/fmesh/port"
	"runtime/debug"
)

// @TODO: add a builder pattern implementation
type Component struct {
	Name           string
	Description    string
	Inputs         port.Ports
	Outputs        port.Ports
	ActivationFunc func(inputs port.Ports, outputs port.Ports) error
}

type Components []*Component

func (c *Component) Activate() (aRes hop.ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			aRes = hop.ActivationResult{
				Activated:     true,
				ComponentName: c.Name,
				Err:           fmt.Errorf("panicked with %w, stacktrace: %s", r, debug.Stack()),
			}
		}
	}()

	//@TODO:: https://github.com/hovsep/fmesh/issues/15
	if !c.Inputs.AnyHasSignal() {
		//No Inputs set, stop here

		aRes = hop.ActivationResult{
			Activated:     false,
			ComponentName: c.Name,
			Err:           nil,
		}

		return
	}

	//Run the computation
	err := c.ActivationFunc(c.Inputs, c.Outputs)

	if IsWaitingForInputError(err) {
		aRes = hop.ActivationResult{
			Activated:     false,
			ComponentName: c.Name,
			Err:           nil,
		}

		if !errors.Is(err, ErrWaitingForInputKeepInputs) {
			c.Inputs.ClearSignal()
		}

		return
	}

	//Clear Inputs
	c.Inputs.ClearSignal()

	if err != nil {
		aRes = hop.ActivationResult{
			Activated:     true,
			ComponentName: c.Name,
			Err:           fmt.Errorf("failed to activate component: %w", err),
		}

		return
	}

	aRes = hop.ActivationResult{
		Activated:     true,
		ComponentName: c.Name,
		Err:           nil,
	}

	return
}

func (c *Component) FlushOutputs() {
	for _, out := range c.Outputs {
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

func (components Components) ByName(name string) *Component {
	for _, c := range components {
		if c.Name == name {
			return c
		}
	}
	return nil
}
