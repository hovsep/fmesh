package main

type Component struct {
	Name   string
	input  int
	h      func(input int) int
	output int
}

func (c *Component) getOutput() int {
	return c.output
}

func (c *Component) hasInput() bool {
	return c.input != 0
}

func (c *Component) setInput(input int) {
	c.input = input
}

func (c *Component) setOutput(output int) {
	c.output = output
}

func (c *Component) clearInput() {
	c.input = 0
}

func (c *Component) clearOutput() {
	c.output = 0
}

func (c *Component) compute() {
	if !c.hasInput() {
		return
	}
	c.output = c.h(c.input)
	c.clearInput()
}
