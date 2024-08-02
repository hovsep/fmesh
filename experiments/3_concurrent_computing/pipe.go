package main

type Pipe struct {
	From *Port
	To   *Port
}

type Pipes []*Pipe

func (p *Pipe) flush() {
	p.To.setValue(p.From.getValue())
	p.From.setValue(nil)
}
