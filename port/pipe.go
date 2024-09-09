package port

type Pipe struct {
	From *Port
	To   *Port
}

type Pipes []*Pipe

func (p *Pipe) Flush() {
	ForwardSignal(p.From, p.To)
}
