package main

type FMesh struct {
	Components Components
	Pipes      Pipes
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

func (fm *FMesh) run() {
	for fm.hasComponentToCompute() {
		fm.compute()
		fm.flush()
	}
}

func (fm *FMesh) hasComponentToCompute() bool {
	for _, c := range fm.Components {
		if c.inputs.anyHasValue() {
			return true
		}
	}
	return false
}
