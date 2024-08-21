package main

import (
	"errors"
	"fmt"
)

type Component struct {
	name    string
	inputs  Ports
	outputs Ports
	handler func(inputs Ports, outputs Ports) error
}

type Components []*Component

func (c *Component) activate() (aRes ActivationResult) {
	defer func() {
		if r := recover(); r != nil {
			aRes = ActivationResult{
				activated:     true,
				componentName: c.name,
				err:           errors.New("component panicked"),
			}
		}
	}()

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
		if !out.hasValue() || len(out.pipes) == 0 {
			continue
		}

		for _, pipe := range out.pipes {
			//Multiplexing
			pipe.To.setValue(out.getValue())
		}
		out.clearValue()
	}
}
