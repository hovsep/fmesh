package main

type Pipe struct {
	In  *Component
	Out *Component
}

func (p *Pipe) flush() {
	p.Out.setInput(p.In.getOutput())
	p.In.clearOutput()
}
