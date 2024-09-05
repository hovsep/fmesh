package main

type Pipe struct {
	From *Port
	To   *Port
}

type Pipes []*Pipe

func (p *Pipe) flush() {
	forwardSignal(p.From, p.To)
}
