package main

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// @TODO: add a builder pattern implementation
type Component struct {
	name        string
	description string
	inputs      Ports
	outputs     Ports
	handler     func(inputs Ports, outputs Ports) error
}

func (c *Component) activate() (aRes ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			aRes = ActivationResult{
				activated:     true,
				componentName: c.name,
				err:           fmt.Errorf("panicked with %w, stacktrace: %s", r, debug.Stack()),
			}
		}
	}()

	//@TODO:: https://github.com/hovsep/fmesh/issues/15
	if !c.inputs.anyHasValue() {
		//No inputs set, stop here

		aRes = ActivationResult{
			activated:     false,
			componentName: c.name,
			err:           nil,
		}

		return
	}

	//Run the computation
	err := c.handler(c.inputs, c.outputs)

	if isWaitingForInput(err) {
		aRes = ActivationResult{
			activated:     false,
			componentName: c.name,
			err:           nil,
		}

		if !errors.Is(err, errWaitingForInputKeepInputs) {
			c.inputs.clearAll()
		}

		return
	}

	//Clear inputs
	c.inputs.clearAll()

	if err != nil {
		aRes = ActivationResult{
			activated:     true,
			componentName: c.name,
			err:           fmt.Errorf("failed to activate component: %w", err),
		}

		return
	}

	aRes = ActivationResult{
		activated:     true,
		componentName: c.name,
		err:           nil,
	}

	return
}

func (c *Component) flushOutputs() {
	for _, out := range c.outputs {
		if !out.hasSignal() || len(out.pipes) == 0 {
			continue
		}

		for _, pipe := range out.pipes {
			//Multiplexing
			pipe.flush()
		}
		out.clearSignal()
	}
}
