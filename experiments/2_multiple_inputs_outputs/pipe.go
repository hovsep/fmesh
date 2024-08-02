package main

type Pipe struct {
	In  *Port
	Out *Port
}

type Pipes []*Pipe

func (p *Pipe) flush() {
	p.Out.setValue(p.In.getValue())
	p.In.setValue(nil)
}
