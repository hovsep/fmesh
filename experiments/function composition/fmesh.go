package main

import "fmt"

type FMesh struct {
	Components []*Component
	Pipes      []*Pipe
}

func (fm *FMesh) compute() {
	for _, c := range fm.Components {
		c.compute()
	}
}

func (fm *FMesh) flush() {
	for _, p := range fm.Pipes {
		p.flush()
	}
}

func (fm *FMesh) hasComponentToCompute() bool {
	for _, c := range fm.Components {
		if c.hasInput() {
			return true
		}
	}
	return false
}

func RunAsFMesh(input int) {
	c1 := Component{
		Name: "mul 2",
	}
	c1.h = func(input int) int {
		return Mul(input, 2)
	}
	c1.setInput(10)

	c2 := Component{
		Name: "add 3",
	}
	c2.h = func(input int) int {
		return Add(input, 3)
	}

	c3 := Component{
		Name: "add 5",
	}
	c3.h = func(input int) int {
		return Add(input, 5)
	}

	fmesh := &FMesh{
		Components: []*Component{&c1, &c2, &c3},
		Pipes: []*Pipe{
			&Pipe{
				In:  &c1,
				Out: &c2,
			},
			&Pipe{
				In:  &c2,
				Out: &c3,
			},
		},
	}

	for fmesh.hasComponentToCompute() {
		fmesh.compute()
		fmesh.flush()
	}

	res := c3.getOutput()
	fmt.Printf("Result is %v", res)
}
