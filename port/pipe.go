package port

//@TODO:the pipe type can be potentially removed

// Pipe is the connection between two ports
type Pipe struct {
	From *Port
	To   *Port
}

// Pipes is a useful collection type
type Pipes []*Pipe

// NewPipe returns new pipe
func NewPipe(from *Port, to *Port) *Pipe {
	return &Pipe{
		From: from,
		To:   to,
	}
}

// Flush makes the signals flow from "From" to "To" port (From is not cleared)
func (p *Pipe) Flush() {
	ForwardSignal(p.From, p.To)
}
